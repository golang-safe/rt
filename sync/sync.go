// Package sync is SafeGo's mutual exclusion.
//
// The host implementation is a standard mutex so that shared-state logic is testable
// under `go test -race` (D1). On the target it lowers to the immediate ceiling priority
// protocol: Lock raises the caller to the mutex's ceiling — the maximum base priority
// of any task that can reach a Lock on it, computed at compile time from the call
// graph — and Unlock restores what Lock saved.
//
// The consequence is worth stating plainly, because it is why this type has no
// TryLock and no timeout: a task holding the mutex runs at the ceiling, so no task that
// could contend is schedulable. Lock never blocks, and deadlock is impossible by
// construction. A TryLock would be an operation that can only ever succeed.
package sync

import (
	stdsync "sync"
)

// Mutex protects shared state between tasks.
//
// A Mutex may only be created during the init phase (D4), which is checked statically,
// and must be locked and unlocked on every path — an unbalanced Lock is rejected at
// compile time rather than left to deadlock at run time.
//
// The zero value is not usable: New returns the value, so that creation is a call the
// phase analysis can see.
type Mutex struct {
	m stdsync.Mutex
}

// New returns a mutex. It may only be called during the init phase.
func New() *Mutex {
	return &Mutex{}
}

// Lock acquires the mutex.
//
// On the target this raises the caller to the mutex's ceiling and returns; it does not
// wait, because nothing that could contend can be running.
func (m *Mutex) Lock() {
	m.m.Lock()
}

// Unlock releases the mutex and restores the priority Lock saved.
func (m *Mutex) Unlock() {
	m.m.Unlock()
}
