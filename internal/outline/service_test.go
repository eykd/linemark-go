package outline

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/eykd/linemark-go/internal/lock"
)

// mockLocker is a test double for the Locker interface.
type mockLocker struct {
	tryLockErr    error
	unlockErr     error
	tryLockCalled bool
	unlockCalled  bool
}

func (m *mockLocker) TryLock(ctx context.Context) error {
	m.tryLockCalled = true
	return m.tryLockErr
}

func (m *mockLocker) Unlock() error {
	m.unlockCalled = true
	return m.unlockErr
}

func TestOutlineService_MutatingCommands_Locking(t *testing.T) {
	type mutatingFn func(context.Context, string, string) (*ModifyResult, error)

	commands := []struct {
		name string
		call func(*OutlineService) mutatingFn
	}{
		{"AddType", func(s *OutlineService) mutatingFn { return s.AddType }},
		{"RemoveType", func(s *OutlineService) mutatingFn { return s.RemoveType }},
	}

	tests := []struct {
		name       string
		tryLockErr error
		wantErr    bool
		wantErrIs  error
		wantUnlock bool
	}{
		{
			name:       "acquires and releases lock",
			wantUnlock: true,
		},
		{
			name:       "fails fast when already locked",
			tryLockErr: lock.ErrAlreadyLocked,
			wantErr:    true,
			wantErrIs:  lock.ErrAlreadyLocked,
			wantUnlock: false,
		},
		{
			name:       "propagates TryLock error",
			tryLockErr: fmt.Errorf("permission denied"),
			wantErr:    true,
			wantUnlock: false,
		},
	}

	for _, cmd := range commands {
		for _, tt := range tests {
			t.Run(cmd.name+"/"+tt.name, func(t *testing.T) {
				locker := &mockLocker{tryLockErr: tt.tryLockErr}
				svc := NewOutlineService(locker)
				fn := cmd.call(svc)

				_, err := fn(context.Background(), "notes", "001")

				if !locker.tryLockCalled {
					t.Error("should call TryLock before mutating")
				}
				if tt.wantErr {
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
						t.Errorf("error = %v, want %v", err, tt.wantErrIs)
					}
				} else if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if locker.unlockCalled != tt.wantUnlock {
					t.Errorf("unlock called = %v, want %v", locker.unlockCalled, tt.wantUnlock)
				}
			})
		}
	}
}

func TestOutlineService_ReadOnlyCommands_BypassLocking(t *testing.T) {
	type readOnlyFn func(ctx context.Context) error

	commands := []struct {
		name string
		call func(*OutlineService) readOnlyFn
	}{
		{"ListTypes", func(s *OutlineService) readOnlyFn {
			return func(ctx context.Context) error {
				_, err := s.ListTypes(ctx, "001")
				return err
			}
		}},
		{"Check", func(s *OutlineService) readOnlyFn {
			return func(ctx context.Context) error {
				_, err := s.Check(ctx)
				return err
			}
		}},
	}

	tests := []struct {
		name       string
		tryLockErr error
	}{
		{"succeeds without lock contention", nil},
		{"succeeds with lock pre-acquired", lock.ErrAlreadyLocked},
	}

	for _, cmd := range commands {
		for _, tt := range tests {
			t.Run(cmd.name+"/"+tt.name, func(t *testing.T) {
				locker := &mockLocker{tryLockErr: tt.tryLockErr}
				svc := NewOutlineService(locker)
				fn := cmd.call(svc)

				err := fn(context.Background())

				if err != nil {
					t.Errorf("read-only command should succeed, got: %v", err)
				}
				if locker.tryLockCalled {
					t.Error("read-only command should not call TryLock")
				}
				if locker.unlockCalled {
					t.Error("read-only command should not call Unlock")
				}
			})
		}
	}
}

func TestOutlineService_Repair_Locking(t *testing.T) {
	tests := []struct {
		name       string
		tryLockErr error
		wantErr    bool
		wantErrIs  error
		wantUnlock bool
	}{
		{
			name:       "acquires and releases lock",
			wantUnlock: true,
		},
		{
			name:       "fails fast when already locked",
			tryLockErr: lock.ErrAlreadyLocked,
			wantErr:    true,
			wantErrIs:  lock.ErrAlreadyLocked,
			wantUnlock: false,
		},
		{
			name:       "propagates TryLock error",
			tryLockErr: fmt.Errorf("permission denied"),
			wantErr:    true,
			wantUnlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker := &mockLocker{tryLockErr: tt.tryLockErr}
			svc := NewOutlineService(locker)

			_, err := svc.Repair(context.Background())

			if !locker.tryLockCalled {
				t.Error("Repair should call TryLock before mutating")
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if locker.unlockCalled != tt.wantUnlock {
				t.Errorf("unlock called = %v, want %v", locker.unlockCalled, tt.wantUnlock)
			}
		})
	}
}

func TestOutlineService_ErrAlreadyLocked_HasClearMessage(t *testing.T) {
	locker := &mockLocker{tryLockErr: lock.ErrAlreadyLocked}
	svc := NewOutlineService(locker)

	_, err := svc.AddType(context.Background(), "notes", "001")
	if err == nil {
		t.Fatal("expected error")
	}
	want := "another lmk command is already running"
	if err.Error() != want {
		t.Errorf("error message = %q, want %q", err.Error(), want)
	}
}
