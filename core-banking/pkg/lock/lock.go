package lock

import (
	"context"
	"fmt"
	"sync"
)

type LockManager interface {
	Lock(ctx context.Context, key string) error
	Unlock(key string)
}

type AccountLockManager struct {
	locks sync.Map
}

func NewAccountLockManager() *AccountLockManager {
	return &AccountLockManager{}
}

func (m *AccountLockManager) getLock(key string) *sync.Mutex {
	val, _ := m.locks.LoadOrStore(key, &sync.Mutex{})
	return val.(*sync.Mutex)
}

func (m *AccountLockManager) Lock(ctx context.Context, key string) error {
	lock := m.getLock(key)

	ch := make(chan struct{}, 1)

	go func() {
		lock.Lock()
		ch <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("lock timeout")
	case <-ch:
		return nil
	}
}

func (m *AccountLockManager) Unlock(key string) {
	lock := m.getLock(key)
	lock.Unlock()
}

func BuildOrderedKey(a, b string) string {
	if a < b {
		return a + ":" + b
	}
	return b + ":" + a
}
