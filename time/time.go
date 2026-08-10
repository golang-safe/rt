// Package time is SafeGo's time source.
//
// It exists so that the same source runs on the host and on the target (D1). Here the
// implementation is backed by the standard library, which makes a control loop
// testable, benchmarkable and fuzzable off-target. The transpiler does not translate
// this code: it recognises the API and lowers each operation to the runtime's tickless
// time core (§2.2.8), so nothing in this file reaches a device.
//
// The standard library is imported here and nowhere else in a SafeGo program. That is
// the point of the rt/ boundary: a whitelist of packages whose host implementation and
// target lowering are maintained together (§2.1.2).
package time

import (
	stdtime "time"
)

// Duration is a span of time in nanoseconds.
//
// It is int64 rather than the platform int so that host and target agree (§2.2.1), and
// nanoseconds rather than counter units so that a program reads the same on every
// target. The conversion to counter units folds to a multiply and a shift, because the
// counter frequency is a compile-time constant from the target descriptor.
type Duration int64

// The familiar constants, so that 100 * time.Millisecond reads as expected.
const (
	Nanosecond  Duration = 1
	Microsecond          = 1000 * Nanosecond
	Millisecond          = 1000 * Microsecond
	Second               = 1000 * Millisecond
)

// Instant is a reading of the monotonic counter.
//
// It is deliberately not a wall clock: there is no calendar on a bare-metal target, and
// a monotonic reading is what timing analysis and deadline arithmetic need. Instants
// are comparable and subtractable; they cannot be constructed from a number.
type Instant struct {
	ns int64
}

// Now returns the current counter reading.
func Now() Instant {
	return Instant{ns: stdtime.Now().UnixNano()}
}

// Add returns the instant d after i.
func (i Instant) Add(d Duration) Instant {
	return Instant{ns: i.ns + int64(d)}
}

// Sub returns the span from earlier to i.
func (i Instant) Sub(earlier Instant) Duration {
	return Duration(i.ns - earlier.ns)
}

// Before reports whether i is earlier than other.
func (i Instant) Before(other Instant) bool {
	return i.ns < other.ns
}

// After reports whether i is later than other.
func (i Instant) After(other Instant) bool {
	return i.ns > other.ns
}

// Sleep blocks the calling task until d has elapsed.
//
// The task is descheduled rather than spinning: the deadline enters the runtime's
// queue and the task is woken by the compare interrupt.
func Sleep(d Duration) {
	if d <= 0 {
		return
	}

	stdtime.Sleep(stdtime.Duration(d))
}

// Ticker delivers a periodic event without drift.
//
// Each deadline is computed from the previous deadline rather than from the time the
// task happened to wake, so a periodic task does not accumulate wake-up latency. If a
// period is missed because the task did not keep up, the missed ticks are skipped and
// counted rather than delivered as a backlog burst, which would deepen the overload.
//
// A Ticker may only be created during the init phase (D4), which is checked
// statically.
type Ticker struct {
	period   Duration
	next     stdtime.Time
	overruns uint32
}

// NewTicker returns a ticker with the given period, anchored at the time of the call.
func NewTicker(period Duration) *Ticker {
	return &Ticker{
		period: period,
		next:   stdtime.Now().Add(stdtime.Duration(period)),
	}
}

// Wait blocks until the next period. It returns after skipping any periods that have
// already passed, so a late task resumes on the grid rather than working through a
// backlog.
func (t *Ticker) Wait() {
	if t.period <= 0 {
		return
	}

	now := stdtime.Now()

	if now.Before(t.next) {
		stdtime.Sleep(t.next.Sub(now))
	}

	t.next = t.next.Add(stdtime.Duration(t.period))

	for !t.next.After(now) {
		t.next = t.next.Add(stdtime.Duration(t.period))
		t.overruns++
	}
}

// Overruns returns how many periods have been skipped.
//
// This is what lets a control loop notice it is not meeting its period instead of
// silently running late, and it is why the counter is part of the API rather than an
// internal statistic.
func (t *Ticker) Overruns() uint32 {
	return t.overruns
}
