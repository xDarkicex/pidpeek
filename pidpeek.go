// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

// Get retrieves process metrics for the given PID.
func Get(pid int) (Metrics, error) {
	return DefaultInspector.Get(pid)
}

// GetIdentity retrieves process identity for the given PID.
func GetIdentity(pid int) (Identity, error) {
	return DefaultInspector.GetIdentity(pid)
}

// GetAll retrieves both metrics and identity for the given PID efficiently.
func GetAll(pid int) (ProcessInfo, error) {
	return DefaultInspector.GetAll(pid)
}

// Self retrieves metrics for the current process.
func Self() (Metrics, error) {
	return DefaultInspector.Self()
}

// SelfIdentity retrieves identity for the current process.
func SelfIdentity() (Identity, error) {
	return DefaultInspector.SelfIdentity()
}
