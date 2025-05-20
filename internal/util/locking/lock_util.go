package locking

import (
	"context"
	"sync"
	"time"
)

// LockUtil provides utility functions for locking
type LockUtil struct {
	locks sync.Map
}

// NewLockUtil creates a new LockUtil
func NewLockUtil() *LockUtil {
	return &LockUtil{}
}

// Lock represents a lock
type Lock struct {
	mu      sync.Mutex
	locked  bool
	timeout time.Duration
}

// NewLock creates a new Lock
func (l *LockUtil) NewLock(timeout time.Duration) *Lock {
	return &Lock{
		timeout: timeout,
	}
}

// Acquire acquires the lock
func (l *Lock) Acquire(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.locked {
		return ErrLockAlreadyAcquired
	}

	if l.timeout > 0 {
		timer := time.NewTimer(l.timeout)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return ErrLockTimeout
		default:
			l.locked = true
			return nil
		}
	}

	l.locked = true
	return nil
}

// Release releases the lock
func (l *Lock) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.locked {
		return ErrLockNotAcquired
	}

	l.locked = false
	return nil
}

// IsLocked checks if the lock is locked
func (l *Lock) IsLocked() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.locked
}

// GetLock gets a lock for a key
func (l *LockUtil) GetLock(key string) *Lock {
	value, _ := l.locks.LoadOrStore(key, l.NewLock(0))
	return value.(*Lock)
}

// ReleaseLock releases a lock for a key
func (l *LockUtil) ReleaseLock(key string) error {
	if value, ok := l.locks.Load(key); ok {
		lock := value.(*Lock)
		if err := lock.Release(); err != nil {
			return err
		}
		l.locks.Delete(key)
		return nil
	}
	return ErrLockNotAcquired
}

// IsLocked checks if a key is locked
func (l *LockUtil) IsLocked(key string) bool {
	if value, ok := l.locks.Load(key); ok {
		return value.(*Lock).IsLocked()
	}
	return false
}

// WithLock executes a function with a lock
func (l *LockUtil) WithLock(ctx context.Context, key string, fn func() error) error {
	lock := l.GetLock(key)
	if err := lock.Acquire(ctx); err != nil {
		return err
	}
	defer lock.Release()
	return fn()
}

// WithTimeoutLock executes a function with a lock and timeout
func (l *LockUtil) WithTimeoutLock(ctx context.Context, key string, timeout time.Duration, fn func() error) error {
	lock := l.NewLock(timeout)
	if err := lock.Acquire(ctx); err != nil {
		return err
	}
	defer lock.Release()
	return fn()
}

// Errors
var (
	ErrLockAlreadyAcquired = &LockError{msg: "lock already acquired"}
	ErrLockNotAcquired     = &LockError{msg: "lock not acquired"}
	ErrLockTimeout         = &LockError{msg: "lock timeout"}
)

// LockError represents a lock error
type LockError struct {
	msg string
}

// Error implements the error interface
func (e *LockError) Error() string {
	return e.msg
}
