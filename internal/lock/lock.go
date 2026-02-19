// Package lock provides advisory file locking for mutating commands.
package lock

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/flock"
)

// ErrAlreadyLocked is returned when another lmk process holds the lock.
var ErrAlreadyLocked = errors.New("another lmk command is already running")

// Flocker abstracts the subset of flock.Flock used for advisory locking.
type Flocker interface {
	TryLock() (bool, error)
	Unlock() error
}

// Lock wraps a Flocker to provide fail-fast advisory locking.
type Lock struct {
	flocker Flocker
}

// New creates a Lock from the given Flocker.
func New(f Flocker) *Lock {
	return &Lock{flocker: f}
}

// NewFromPath creates a Lock backed by a file at the given path.
func NewFromPath(path string) *Lock {
	return New(flock.New(path))
}

// TryLock attempts a non-blocking lock acquisition. It returns
// ErrAlreadyLocked if the lock is held by another process, or wraps
// any underlying error from the Flocker.
func (l *Lock) TryLock(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ok, err := l.flocker.TryLock()
	if err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	if !ok {
		return ErrAlreadyLocked
	}
	return nil
}

// Unlock releases the advisory lock.
func (l *Lock) Unlock() error {
	if err := l.flocker.Unlock(); err != nil {
		return fmt.Errorf("releasing lock: %w", err)
	}
	return nil
}
