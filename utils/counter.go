package utils

import "sync"

type AtomicInt32Counter struct {
	counter int32
	lock    sync.RWMutex
}

func (a *AtomicInt32Counter) Increment(delta int32) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.counter += delta
}

func (a *AtomicInt32Counter) Count() int32 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.counter
}
