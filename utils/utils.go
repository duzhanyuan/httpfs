package utils

import (
	"io"
	"os"
	"strconv"

	"github.com/prologic/httpfs/types"
)

// ReadDir ...
func ReadDir(dirname string) ([]types.Entry, error) {
	f, err := os.Open(dirname)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	xs, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	entries := make([]types.Entry, len(xs))
	for i, x := range xs {
		entries[i] = types.Entry{
			Name:    x.Name(),
			Size:    x.Size(),
			Mode:    uint32(x.Mode()),
			ModTime: x.ModTime().UTC().Unix(),
			IsDir:   x.IsDir(),
		}
	}

	return entries, nil
}

// FileSize return the size of the open file
func FileSize(f *os.File) (int64, error) {
	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// SafeParseBool ...
func SafeParseBool(s string, d bool) bool {
	b, e := strconv.ParseBool(s)
	if e != nil {
		return d
	}
	return b
}

// SafeParseInt ...
func SafeParseInt(s string, d int) int {
	n, e := strconv.ParseInt(s, 10, 32)
	if e != nil {
		return d
	}
	return int(n)
}

// SafeParseInt64 ...
func SafeParseInt64(s string, d int64) int64 {
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil {
		return d
	}
	return n
}

// SafeStatSize ...
func SafeStatSize(path string) int64 {
	d, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return d.Size()
}
