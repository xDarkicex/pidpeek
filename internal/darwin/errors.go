// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import "errors"

// Sentinel errors for darwin package.
var (
	ErrProcessNotFound = errors.New("process not found")
	ErrAccessDenied    = errors.New("access denied")
	ErrResourceExhausted = errors.New("resource exhausted")
)