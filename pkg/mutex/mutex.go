package mutex

import (
	"sync"
	"sync/atomic"
)

type ObservableMutex struct {
	m        sync.Mutex
	reserved atomic.Bool
}

func NewObservableMutex() *ObservableMutex {
	return &ObservableMutex{
		m:        sync.Mutex{},
		reserved: atomic.Bool{},
	}
}

func (r *ObservableMutex) Lock() {
	r.m.Lock()
	r.reserved.Store(true)
}

func (r *ObservableMutex) Unlock() {
	r.reserved.Store(false)
	r.m.Unlock()
}

// IsLocked returns whether the mutex is locked
func (r *ObservableMutex) IsLocked() bool {
	return r.reserved.Load()
}

func (r *ObservableMutex) ReservedFunc(fn func() error) error {
	r.Lock()
	defer r.Unlock()
	return fn()
}

func (r *ObservableMutex) IfUnlocked(fn func() error) error {
	if !r.m.TryLock() {
		return nil
	}

	defer func() {
		r.reserved.Store(false)
		r.m.Unlock()
	}()

	return fn()
}
