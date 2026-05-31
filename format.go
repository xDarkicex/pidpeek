// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

// FormatBytes formats bytes as IEC binary: KiB, MiB, GiB.
func FormatBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for n >= unit*unit && exp < 4 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGT"[exp])
}