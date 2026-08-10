// Package errors is the SafeGo error intrinsic.
//
// An error is a code and a description, and nothing else (D12, §2.2.6). Custom error types
// are rejected by the subset, so there is exactly one representation: two words, returned by
// value, with no allocation and no wrapping chain. Code zero means no error, which is what
// makes the zero value equal to nil.
//
// The transpiler does not lower calls into this package — it recognises them. New folds to the
// value it denotes and Is folds to an integer comparison, so error handling on target costs
// the same as comparing two ints. This file is the host implementation of the same semantics,
// which is what lets a program that uses errors run under `go test` (D1).
package errors

// Error is the canonical error value.
//
// It is a value type with unexported fields: no one can construct a variant of it, and copying
// it copies everything it is. Both properties are what the single-representation rule needs.
type Error struct {
	code int32
	desc string
}

// New returns the error with a code and a description.
//
// Both arguments must be constants in transpiled code: the description lives in .rodata, and a
// runtime-built string would need storage that no task owns. The host signature does not
// enforce that — the transpiler does, with a diagnostic naming the argument.
func New(code int32, desc string) error {
	return Error{code: code, desc: desc}
}

// Error returns the description, satisfying the error interface on the host.
func (e Error) Error() string {
	return e.desc
}

// Code returns the error's code. It is the identity of an error: two errors with the same code
// are the same error whatever their text.
func (e Error) Code() int32 {
	return e.code
}

// Is reports whether two errors are the same error, by code.
//
// There is no unwrapping. A wrapping chain would need either allocation or a bounded-depth
// convention, and it would make the cost of comparing two errors depend on how deeply one of
// them had been wrapped — which a WCET argument cannot accept.
func Is(err, target error) bool {
	return code(err) == code(target)
}

// code extracts an error's code, treating anything that is not one of ours as unknown but
// non-nil. On the host a foreign error can reach here through an interface; on target it
// cannot, because no other type satisfies error.
func code(err error) int32 {
	if err == nil {
		return 0
	}

	// A direct type assertion, not errors.As: there is no wrapping in this package, and
	// adding unwrapping support would make comparing two errors cost more the more deeply one
	// had been wrapped.
	if e, ok := err.(Error); ok { //nolint:errorlint // There is no wrapping to see through.
		return e.code
	}

	return -1
}
