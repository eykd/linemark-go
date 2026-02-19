package lock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/eykd/linemark-go/internal/lock"
)

// mockFlocker is a test double for the Flocker interface.
type mockFlocker struct {
	tryLockResult bool
	tryLockErr    error
	unlockErr     error
	tryLockCalled bool
	unlockCalled  bool
}

func (m *mockFlocker) TryLock() (bool, error) {
	m.tryLockCalled = true
	return m.tryLockResult, m.tryLockErr
}

func (m *mockFlocker) Unlock() error {
	m.unlockCalled = true
	return m.unlockErr
}

func TestLock_TryLock(t *testing.T) {
	errPermDenied := errors.New("permission denied")

	tests := []struct {
		name          string
		tryLockResult bool
		tryLockErr    error
		wantErr       error
	}{
		{
			name:          "succeeds when lock is available",
			tryLockResult: true,
			wantErr:       nil,
		},
		{
			name:          "returns ErrAlreadyLocked when lock is held",
			tryLockResult: false,
			wantErr:       lock.ErrAlreadyLocked,
		},
		{
			name:          "wraps underlying flock error",
			tryLockResult: false,
			tryLockErr:    errPermDenied,
			wantErr:       errPermDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockFlocker{
				tryLockResult: tt.tryLockResult,
				tryLockErr:    tt.tryLockErr,
			}
			l := lock.New(m)

			err := l.TryLock(context.Background())

			if !m.tryLockCalled {
				t.Error("expected TryLock to be called on flocker")
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLock_TryLock_AlreadyLocked_HasClearMessage(t *testing.T) {
	m := &mockFlocker{
		tryLockResult: false,
		tryLockErr:    nil,
	}
	l := lock.New(m)

	err := l.TryLock(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := "another lmk command is already running"
	if err.Error() != want {
		t.Errorf("error message = %q, want %q", err.Error(), want)
	}
}

func TestLock_Unlock(t *testing.T) {
	tests := []struct {
		name      string
		unlockErr error
		wantErr   bool
	}{
		{
			name:      "succeeds when unlock works",
			unlockErr: nil,
			wantErr:   false,
		},
		{
			name:      "propagates unlock error",
			unlockErr: errors.New("unlock failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockFlocker{
				unlockErr: tt.unlockErr,
			}
			l := lock.New(m)

			err := l.Unlock()

			if !m.unlockCalled {
				t.Error("expected Unlock to be called on flocker")
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.unlockErr != nil && err != nil {
				if !errors.Is(err, tt.unlockErr) {
					t.Errorf("error should wrap %v, got: %v", tt.unlockErr, err)
				}
			}
		})
	}
}

func TestLock_DeferUnlock_ReleasesOnSuccess(t *testing.T) {
	m := &mockFlocker{
		tryLockResult: true,
	}
	l := lock.New(m)

	// Simulate the defer pattern used by mutating commands:
	// acquire lock, defer unlock, do work, return
	func() {
		err := l.TryLock(context.Background())
		if err != nil {
			t.Fatalf("unexpected TryLock error: %v", err)
		}
		defer func() {
			unlockErr := l.Unlock()
			if unlockErr != nil {
				t.Fatalf("unexpected Unlock error: %v", unlockErr)
			}
		}()

		// Simulate successful work (no error)
	}()

	if !m.unlockCalled {
		t.Error("Unlock was not called; defer pattern must release lock on success")
	}
}

func TestLock_DeferUnlock_ReleasesOnError(t *testing.T) {
	m := &mockFlocker{
		tryLockResult: true,
	}
	l := lock.New(m)

	// Simulate the defer pattern when an error occurs mid-operation
	simulatedErr := errors.New("operation failed")
	var capturedErr error

	func() {
		err := l.TryLock(context.Background())
		if err != nil {
			t.Fatalf("unexpected TryLock error: %v", err)
		}
		defer func() {
			unlockErr := l.Unlock()
			if unlockErr != nil {
				t.Fatalf("unexpected Unlock error: %v", unlockErr)
			}
		}()

		// Simulate error mid-operation
		capturedErr = simulatedErr
	}()

	if !m.unlockCalled {
		t.Error("Unlock was not called; defer pattern must release lock on error")
	}
	if capturedErr != simulatedErr {
		t.Error("simulated error was not captured")
	}
}

func TestLock_TryLock_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before TryLock

	m := &mockFlocker{
		tryLockResult: true,
	}
	l := lock.New(m)

	err := l.TryLock(ctx)

	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error should be context.Canceled, got: %v", err)
	}
}

func TestNewFromPath_CreatesLockAtPath(t *testing.T) {
	lockPath := t.TempDir() + "/.linemark/lock"

	l := lock.NewFromPath(lockPath)

	if l == nil {
		t.Fatal("NewFromPath returned nil")
	}
}
