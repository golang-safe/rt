// Package volatile is SafeGo's memory-mapped I/O.
//
// A register access on the target must be a volatile load or store of exactly the declared
// width — never wider, never narrower, never elided, never reordered with its neighbours —
// because a peripheral notices the difference and memory does not. On the host the same API
// is backed by an injectable register bank, so logic that drives hardware is ordinary
// testable Go: it can be unit tested, fuzzed and run under the race detector without a
// board.
//
// **This package needs no `unsafe`, and that is not an accident.** On the target it is not
// Go at all — the transpiler recognises the API and emits the access — and on the host it
// models the address space rather than touching it. Nothing is ever converted to a pointer.
//
// **The accessors are functions rather than variables for a reason that only shows on the
// host.** A package-level `uint32` would let a test set a register and would lower to a
// volatile access easily enough, but the Go compiler may keep it in a register across a
// loop — so a poll such as `for !status.HasBits(ready) {}` would spin forever on the host
// while working on the target. A call cannot be cached that way, so both halves agree.
//
// That is the concrete payoff the whole design was argued for. A driver's decisions — the
// bit patterns, the ordering, the state machine — are exactly the part most likely to be
// wrong and least likely to be exercised on hardware, and this is what lets them be tested
// where testing is cheap.
//
// Like everything under rt/, this is an intrinsic: the transpiler recognises the API and
// emits the access directly. Nothing in this file is translated, which is why it may use
// the standard library the subset otherwise bans.
package volatile

import "sync"

// A register is its address. The types are distinct so that the *width* of an access is
// part of the program's text rather than a property of whatever integer happens to be
// assigned — a 32-bit register read as 16 bits is a bug no type checker would otherwise
// catch, and on many buses it faults.
type (
	// Reg8 is an 8-bit memory-mapped register.
	Reg8 uintptr

	// Reg16 is a 16-bit memory-mapped register.
	Reg16 uintptr

	// Reg32 is a 32-bit memory-mapped register.
	Reg32 uintptr

	// Reg64 is a 64-bit memory-mapped register.
	Reg64 uintptr
)

// bank is the host's stand-in for the address space.
//
// It is a map rather than a block of memory because a register is not memory: reading one
// that was never written is a question about a device, and answering zero is a decision the
// test should make rather than one this package makes silently. Tests set what they mean to
// model.
var bank struct {
	sync.RWMutex

	values map[uintptr]uint64
}

func init() {
	bank.values = map[uintptr]uint64{}
}

// Reset clears the modelled register bank. A test that does not start from a known state is
// testing the previous test as well as its own.
func Reset() {
	bank.Lock()
	defer bank.Unlock()

	bank.values = map[uintptr]uint64{}
}

// Peek and Poke read and write the model directly, which is how a test stands in for the
// device: Poke to present a value the code must react to, Peek to see what it wrote.
//
// Neither exists on the target — a program calling them is calling something that models
// hardware rather than touching it, which is the one thing this package must not let be
// confused. The transpiler refuses them.
// Peek reads the modelled register bank directly.
func Peek(address uintptr) uint64 {
	bank.RLock()
	defer bank.RUnlock()

	return bank.values[address]
}

// Poke writes the modelled register bank directly, presenting a value the code under test
// must react to.
func Poke(address uintptr, value uint64) {
	bank.Lock()
	defer bank.Unlock()

	bank.values[address] = value
}

func load(address uintptr) uint64 {
	bank.RLock()
	defer bank.RUnlock()

	return bank.values[address]
}

func store(address uintptr, value uint64) {
	bank.Lock()
	defer bank.Unlock()

	bank.values[address] = value
}

// Load reads the register.
//
// The model stores every width in one place, so a load narrows — and narrowing is right:
// reading an 8-bit register returns eight bits whatever a test happened to poke, which is
// what the hardware does.
func (r Reg8) Load() uint8 { return uint8(load(uintptr(r))) } //nolint:gosec // Narrowing is the access width.

// Store writes the register.
func (r Reg8) Store(v uint8) { store(uintptr(r), uint64(v)) }

// SetBits sets every bit of the mask, leaving the rest as they were.
//
// It is a read-modify-write and is not atomic against an interrupt that touches the same
// register. Where that matters the access belongs in a critical section, which the caller
// owns: hiding one here would put a mask around every register write in the program,
// including the great majority that do not need it.
func (r Reg8) SetBits(mask uint8) { r.Store(r.Load() | mask) }

// ClearBits clears every bit of the mask, leaving the rest as they were.
func (r Reg8) ClearBits(mask uint8) { r.Store(r.Load() &^ mask) }

// HasBits reports whether every bit of the mask is set.
func (r Reg8) HasBits(mask uint8) bool { return (r.Load() & mask) == mask }

// Load reads the register.
func (r Reg16) Load() uint16 { return uint16(load(uintptr(r))) } //nolint:gosec // Narrowing is the access width.

// Store writes the register.
func (r Reg16) Store(v uint16) { store(uintptr(r), uint64(v)) }

// SetBits sets every bit of the mask.
func (r Reg16) SetBits(mask uint16) { r.Store(r.Load() | mask) }

// ClearBits clears every bit of the mask.
func (r Reg16) ClearBits(mask uint16) { r.Store(r.Load() &^ mask) }

// HasBits reports whether every bit of the mask is set.
func (r Reg16) HasBits(mask uint16) bool { return (r.Load() & mask) == mask }

// Load reads the register.
func (r Reg32) Load() uint32 { return uint32(load(uintptr(r))) } //nolint:gosec // Narrowing is the access width.

// Store writes the register.
func (r Reg32) Store(v uint32) { store(uintptr(r), uint64(v)) }

// SetBits sets every bit of the mask.
func (r Reg32) SetBits(mask uint32) { r.Store(r.Load() | mask) }

// ClearBits clears every bit of the mask.
func (r Reg32) ClearBits(mask uint32) { r.Store(r.Load() &^ mask) }

// HasBits reports whether every bit of the mask is set.
func (r Reg32) HasBits(mask uint32) bool { return (r.Load() & mask) == mask }

// Load reads the register.
func (r Reg64) Load() uint64 { return load(uintptr(r)) }

// Store writes the register.
func (r Reg64) Store(v uint64) { store(uintptr(r), v) }

// SetBits sets every bit of the mask.
func (r Reg64) SetBits(mask uint64) { r.Store(r.Load() | mask) }

// ClearBits clears every bit of the mask.
func (r Reg64) ClearBits(mask uint64) { r.Store(r.Load() &^ mask) }

// HasBits reports whether every bit of the mask is set.
func (r Reg64) HasBits(mask uint64) bool { return (r.Load() & mask) == mask }
