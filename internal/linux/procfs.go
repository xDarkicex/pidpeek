//go:build linux

// Package linux provides Linux-specific process metrics via procfs.
package linux

import (
	"bytes"
	"errors"
	"strconv"
)

// StatFields holds parsed fields from /proc/<pid>/stat after the comm.
type StatFields struct {
	State      byte
	Ppid       int32
	Utime      uint64
	Stime      uint64
	NumThreads int32
	Starttime  uint64
	VSize      uint64
}

// ParseComm extracts the comm field from a /proc/<pid>/stat line.
// comm is the portion between the first '(' and last ')'.
func ParseComm(data []byte) (string, error) {
	start := bytes.IndexByte(data, '(')
	end := bytes.LastIndexByte(data, ')')
	if start == -1 || end == -1 || start >= end {
		return "", errors.New("invalid stat format: missing parentheses")
	}
	return string(data[start+1 : end]), nil
}

// ParseStatFields parses the remainder of a /proc/<pid>/stat line after the comm.
// Returns ErrProcessNotFound if fewer than 22 fields are present.
func ParseStatFields(data []byte) (*StatFields, error) {
	lastParen := bytes.LastIndexByte(data, ')')
	if lastParen == -1 {
		return nil, ErrProcessNotFound
	}

	remainder := data[lastParen+2:] // skip ") " prefix
	fields := bytes.Fields(remainder)
	if len(fields) < 22 {
		return nil, ErrProcessNotFound
	}

	ppid, err := strconv.ParseInt(string(fields[1]), 10, 32)
	if err != nil {
		return nil, ErrProcessNotFound
	}
	utime, err := strconv.ParseUint(string(fields[11]), 10, 64)
	if err != nil {
		return nil, ErrProcessNotFound
	}
	stime, err := strconv.ParseUint(string(fields[12]), 10, 64)
	if err != nil {
		return nil, ErrProcessNotFound
	}
	numThreads, err := strconv.ParseInt(string(fields[17]), 10, 32)
	if err != nil {
		return nil, ErrProcessNotFound
	}
	starttime, err := strconv.ParseUint(string(fields[19]), 10, 64)
	if err != nil {
		return nil, ErrProcessNotFound
	}
	vsize, err := strconv.ParseUint(string(fields[20]), 10, 64)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	return &StatFields{
		State:      fields[0][0],
		Ppid:       int32(ppid),
		Utime:      utime,
		Stime:      stime,
		NumThreads: int32(numThreads),
		Starttime:  starttime,
		VSize:      vsize,
	}, nil
}

// ParseStatmPages parses /proc/<pid>/statm and returns the resident pages (field 1).
func ParseStatmPages(data []byte) (uint64, error) {
	fields := bytes.Fields(data)
	if len(fields) < 2 {
		return 0, ErrProcessNotFound
	}
	pages, err := strconv.ParseUint(string(fields[1]), 10, 64)
	if err != nil {
		return 0, ErrProcessNotFound
	}
	return pages, nil
}
