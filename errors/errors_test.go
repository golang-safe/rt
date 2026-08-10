package errors

import "testing"

func TestNilIsCodeZero(t *testing.T) {
	// The zero value being the absence of an error is what lets the emitted C use code zero
	// as nil, with no separate discriminant.
	var zero Error

	if zero.Code() != 0 {
		t.Errorf("the zero Error has code %d, expected 0", zero.Code())
	}

	if !Is(nil, nil) {
		t.Errorf("nil should be Is nil")
	}
}

func TestIsComparesCodes(t *testing.T) {
	first := New(7, "sensor timeout")
	second := New(7, "a different description, same code")
	other := New(8, "sensor timeout")

	if !Is(first, second) {
		t.Errorf("errors with the same code are the same error, whatever their text")
	}

	if Is(first, other) {
		t.Errorf("errors with different codes are different errors, whatever their text")
	}

	if Is(first, nil) {
		t.Errorf("a non-zero error is not nil")
	}
}

func TestErrorReportsItsDescription(t *testing.T) {
	err := New(3, "over temperature")

	if err.Error() != "over temperature" {
		t.Errorf("Error() = %q, expected %q", err.Error(), "over temperature")
	}
}

// TestForeignErrorIsNotAnyOfOurs guards the host-only path: a foreign error reaching Is must
// not compare equal to one of ours, or a host test would pass where the target would differ.
func TestForeignErrorIsNotAnyOfOurs(t *testing.T) {
	if Is(foreign{}, New(0, "")) {
		t.Errorf("a foreign error must not match code zero")
	}

	if Is(foreign{}, nil) {
		t.Errorf("a foreign error is not nil")
	}
}

type foreign struct{}

func (foreign) Error() string { return "foreign" }
