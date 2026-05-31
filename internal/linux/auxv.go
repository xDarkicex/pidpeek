//go:build linux

package linux

import "encoding/binary"

// ReadClkTck reads AT_CLKTCK from an auxv binary blob.
// Each entry is 16 bytes (8-byte type + 8-byte value on 64-bit).
// Returns 100.0 if AT_CLKTCK is not found or the vector ends at AT_NULL first.
func ReadClkTck(data []byte) float64 {
	const entrySize = 16
	const atClkTck = 17
	const atNull = 0

	for offset := 0; offset+entrySize <= len(data); offset += entrySize {
		entry := data[offset : offset+entrySize]
		typ := binary.LittleEndian.Uint64(entry[0:8])
		val := binary.LittleEndian.Uint64(entry[8:16])

		if typ == atNull {
			break
		}
		if typ == atClkTck {
			return float64(val)
		}
	}
	return 100.0
}
