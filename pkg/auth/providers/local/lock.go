package local

import (
	"sync"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/util/lock"
)

// add an extra mux so the same process doesn't try to overwrite the lock
var mux sync.Mutex

func (a *LocalAuthProvider) LockSecret() error {
	mux.Lock()
	a.lock = lock.New(a.client, a.secrets.GetName(), time.Duration(10)*time.Second)
	return a.lock.Acquire()
}

func (a *LocalAuthProvider) ReleaseLock() {
	if err := a.lock.Release(); err != nil {
		localAuthLogger.Error(err, "Failed to release lock")
	}
	a.lock = nil
	mux.Unlock()
}
