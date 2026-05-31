//go:build linux

package linux

import (
	"bytes"
	"strconv"
)

// ParseProcStatBtime finds the btime (boot time in Unix seconds) from /proc/stat.
func ParseProcStatBtime(data []byte) (int64, error) {
	for len(data) > 0 {
		lineEnd := bytes.IndexByte(data, '\n')
		if lineEnd < 0 {
			break
		}
		line := data[:lineEnd]
		data = data[lineEnd+1:]

		if bytes.HasPrefix(line, []byte("btime ")) {
			val, err := strconv.ParseInt(string(bytes.TrimSpace(line[6:])), 10, 64)
			if err != nil {
				return 0, ErrProcessNotFound
			}
			return val, nil
		}
	}
	return 0, ErrProcessNotFound
}
