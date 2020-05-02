package secrets

import (
	"sync"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets/providers/k8secret"
	"github.com/tinyzimmer/kvdi/pkg/util/lock"

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
	backend v1alpha1.SecretsProvider
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
}

// cacheItem is a cached item in the SecretEngine
type cacheItem struct {
	// the contents of the secret
	contents []byte
	// when this cache item expires
	expiresAt time.Time
}

// GetSecretEngine returns a new secret engine for the given cluster. This is
// where variable providers will be checked.
func GetSecretEngine(cluster *v1alpha1.VDICluster) *SecretEngine {
	engine := &SecretEngine{
		backend: k8secret.New(),
		cluster: cluster,
		cache:   make(map[string]*cacheItem),
	}
	return engine
}

// Setup sets the local client inteface and calls Setup on the backend.
func (s *SecretEngine) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	s.client = c
	// rewrite cluster since this is a method that can be used to refresh
	// configuration also.
	s.cluster = cluster
	return s.backend.Setup(c, cluster)
}

// readCache will return the contents of a secret from the cache if still valid.
// Otherwise it returns nil.
func (s *SecretEngine) readCache(name string) []byte {
	if cached, ok := s.cache[name]; ok {
		if cached.expiresAt.Before(time.Now()) {
			return cached.contents
		}
	}
	return nil
}

// writeCache writes a new value to the cache, replacing an existing one of the
// same name.
func (s *SecretEngine) writeCache(name string, contents []byte) {
	s.cache[name] = &cacheItem{
		contents:  contents,
		expiresAt: time.Now().Add(cacheTTL),
	}
}

// ReadSecret will fetch the requested secret from the backend. If cache is true,
// the cache will be checked first, and if not found, then the result of a backend
// query will be written to the cache.
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
	if cache {
		s.writeCache(name, secret)
	}
	return secret, nil
}

// WriteSecret writes the given secret to the backend. If it is also found in
// the cache, then the contents of the value in the cache are replaced with the
// new value.
func (s *SecretEngine) WriteSecret(name string, contents []byte) error {
	if err := s.backend.WriteSecret(name, contents); err != nil {
		return err
	}
	if val := s.readCache(name); val != nil {
		s.writeCache(name, contents)
	}
	return nil
}

// AppendSecret is a convenience wrapper around reading a secret, adding a line,
// and then overwriting the existing secret with the new value. When using this method
// it is assumed to use the cache.
func (s *SecretEngine) AppendSecret(name string, line []byte) error {
	currentVal, err := s.ReadSecret(name, true)
	if err != nil {
		return err
	}
	newLine := append([]byte("\n"), line...)
	newVal := append(currentVal, newLine...)
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
