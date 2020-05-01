package secrets

import (
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets/providers/k8secret"
	"github.com/tinyzimmer/kvdi/pkg/util/lock"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var secretsLog = logf.Log.WithName("secrets")

type SecretEngine struct {
	backend v1alpha1.SecretsProvider
	cluster *v1alpha1.VDICluster
	client  client.Client
	lock    *lock.Lock
	cache   map[string][]byte
}

func GetSecretEngine(cluster *v1alpha1.VDICluster) *SecretEngine {
	engine := &SecretEngine{
		backend: k8secret.New(),
		cluster: cluster,
		cache:   make(map[string][]byte),
	}
	return engine
}

func (s *SecretEngine) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	s.client = c
	return s.backend.Setup(c, cluster)
}

func (s *SecretEngine) ReadSecret(name string, cache bool) ([]byte, error) {
	if cache {
		if val, ok := s.cache[name]; ok {
			return val, nil
		}
	}
	secret, err := s.backend.ReadSecret(name)
	if err != nil {
		return nil, err
	}
	if cache {
		s.cache[name] = secret
	}
	return secret, nil
}

func (s *SecretEngine) WriteSecret(name string, contents []byte) error {
	if err := s.backend.WriteSecret(name, contents); err != nil {
		return err
	}
	if _, ok := s.cache[name]; ok {
		s.cache[name] = contents
	}
	return nil
}

func (s *SecretEngine) Lock() error {
	s.lock = lock.New(s.client, s.cluster.GetAppSecretsName(), time.Duration(10)*time.Second)
	return s.lock.Acquire()
}

func (s *SecretEngine) Release() {
	if err := s.lock.Release(); err != nil {
		secretsLog.Error(err, "Failed to release lock")
	}
	s.lock = nil
}
