package secrets

import (
	"bytes"
	"sync"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/lock"

	"github.com/tinyzimmer/kvdi/pkg/secrets/common"
	"github.com/tinyzimmer/kvdi/pkg/secrets/providers/k8secret"
	"github.com/tinyzimmer/kvdi/pkg/secrets/providers/vault"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// secretsLog is the logr interface for the secrets engine
var secretsLog = logf.Log.WithName("secrets")

// cacheTTL is how long cache items remain valid.
// TODO: make this configurable
var cacheTTL = time.Duration(1) * time.Hour

// SecretEngine is an object wrapper for interacting with backend secret
// "providers". It wraps a cache and a locking mechanism around the simple
// Read/Write methods that the backends provide.
type SecretEngine struct {
	// the provider backend
	backend common.SecretsProvider
	// the cluster configuration
	cluster *v1alpha1.VDICluster
	// the k8s client
	client client.Client
	// the local value cache
	cache map[string]*cacheItem
	// mux for local-process locking
	mux sync.Mutex
	// a pointer used for remote locks
	lock *lock.Lock
	// the ttl on cached items
	cacheTTL time.Duration
}

// cacheItem is a cached item in the SecretEngine
type cacheItem struct {
	// the contents of a bytes secret
	contents []byte
	// the contents of a map secret
	contentsMap map[string][]byte
	// unix timestamp when this cache item expires
	expiresAt int64
}

// GetSecretEngine returns a new secret engine for the given cluster.
func GetSecretEngine(cluster *v1alpha1.VDICluster) *SecretEngine {
	var backend common.SecretsProvider
	switch cluster.GetSecretsBackend() {
	case v1alpha1.SecretsBackendVault:
		backend = vault.New()
	default:
		backend = k8secret.New()
	}
	engine := &SecretEngine{
		backend:  backend,
		cluster:  cluster,
		cache:    make(map[string]*cacheItem),
		cacheTTL: cacheTTL,
	}
	return engine
}

// setClient sets the k8s client for the secret engine.
func (s *SecretEngine) setClient(c client.Client) { s.client = c }

// Setup sets the local client inteface and calls Setup on the backend.
func (s *SecretEngine) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	s.setClient(c)
	if err := s.Lock(); err != nil {
		return err
	}
	defer s.Release()
	// rewrite cluster since this is a method that can be used to refresh
	// configuration also.
	s.cluster = cluster
	return s.backend.Setup(c, cluster)
}

// readCache will return the contents of a secret from the cache if still valid.
// Otherwise it returns nil.
func (s *SecretEngine) readCache(name string) []byte {
	if cached, ok := s.cache[name]; ok {
		if cached.expiresAt > time.Now().Unix() {
			return cached.contents
		}
	}
	return nil
}

// readCacheMap will return the contents of a secret from the cache if still valid.
// Otherwise it returns nil.
func (s *SecretEngine) readCacheMap(name string) map[string][]byte {
	if cached, ok := s.cache[name]; ok {
		if cached.expiresAt > time.Now().Unix() {
			return cached.contentsMap
		}
	}
	return nil
}

// writeCache writes a new bytes value to the cache, replacing an existing one of the
// same name.
func (s *SecretEngine) writeCache(name string, contents []byte) {
	s.cache[name] = &cacheItem{
		contents:  contents,
		expiresAt: time.Now().Add(s.cacheTTL).Unix(),
	}
}

// writeCacheMap writes a new map value to the cache, replacing an existing one of the
// same name.
func (s *SecretEngine) writeCacheMap(name string, contents map[string][]byte) {
	s.cache[name] = &cacheItem{
		contentsMap: contents,
		expiresAt:   time.Now().Add(s.cacheTTL).Unix(),
	}
}

// ReadSecret will fetch the requested secret from the backend. If cache is true,
// the cache will be checked first, and if not found then the backend will be queried.
// The secret is unconditionally written to the cache after retrieval.
func (s *SecretEngine) ReadSecret(name string, cache bool) ([]byte, error) {
	if cache {
		if val := s.readCache(name); val != nil {
			return val, nil
		}
	}
	secret, err := s.backend.ReadSecret(name)
	if err != nil {
		return nil, err
	}
	s.writeCache(name, secret)
	return secret, nil
}

// ReadSecretMap will fetch the requested secret from the backend. If cache is true,
// the cache will be checked first, and if not found the backend will be queried. The result
// is then unconditionally written to the cache.
func (s *SecretEngine) ReadSecretMap(name string, cache bool) (map[string][]byte, error) {
	if cache {
		if val := s.readCacheMap(name); val != nil {
			return val, nil
		}
	}
	secret, err := s.backend.ReadSecretMap(name)
	if err != nil {
		return nil, err
	}
	s.writeCacheMap(name, secret)
	return secret, nil
}

// WriteSecret writes the given secret to the backend. It also unconditionally writes
// it to the local cache.
func (s *SecretEngine) WriteSecret(name string, contents []byte) error {
	if err := s.backend.WriteSecret(name, contents); err != nil {
		return err
	}
	s.writeCache(name, contents)
	return nil
}

// WriteSecretMap writes the given secret map to the backend. It also unconditionally writes
// it to the local cache.
func (s *SecretEngine) WriteSecretMap(name string, contents map[string][]byte) error {
	if err := s.backend.WriteSecretMap(name, contents); err != nil {
		return err
	}
	s.writeCacheMap(name, contents)
	return nil
}

// AppendSecret is a convenience wrapper around reading a secret, adding a line,
// and then overwriting the existing secret with the new value. When using this method
// it is assumed to use the cache.
func (s *SecretEngine) AppendSecret(name string, line []byte) error {
	currentVal, err := s.ReadSecret(name, true)
	if err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return err
		}
		currentVal = make([]byte, 0)
	}
	if len(currentVal) != 0 && !bytes.HasSuffix(currentVal, []byte("\n")) {
		line = append([]byte("\n"), line...)
	}
	newVal := append(currentVal, line...)
	return s.WriteSecret(name, newVal)
}

// Lock locks the secret engine. This is useful for long running operations that
// need to guarantee consistency. If there are multiple replicas of the app running,
// a remote lock is also acquired to keep peer processes from interfering.
func (s *SecretEngine) Lock() error {
	// mux lock to make sure the local process doesn't overwrite the lock
	s.mux.Lock()

	if *s.cluster.GetAppReplicas() > 1 {
		// remote lock to be held against peers
		s.lock = lock.New(s.client, s.cluster.GetAppSecretsName(), time.Duration(10)*time.Second)
		if err := s.lock.Acquire(); err != nil {
			return err
		}
	}
	return nil
}

// Release will release any currently held locks.
func (s *SecretEngine) Release() {
	defer s.mux.Unlock()
	if s.lock != nil {
		if err := s.lock.Release(); err != nil {
			secretsLog.Error(err, "Failed to release lock")
		}
	}
	s.lock = nil
}

// Close calls close on the backend
func (s *SecretEngine) Close() error { return s.backend.Close() }
