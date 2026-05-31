// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

import (
	"errors"
	"fmt"
)

// Sentinel errors used across all platforms.
var (
	ErrProcessNotFound   = errors.New("process not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrResourceExhausted = errors.New("resource exhausted")
)

// wrapErr adds context to a sentinel error.
// All public functions use this; callers can still errors.Is() on the sentinel.
// Produces: "pidpeek.OpenProcess 1234: access denied"
func wrapErr(op string, pid int, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("pidpeek.%s %d: %w", op, pid, err)
}