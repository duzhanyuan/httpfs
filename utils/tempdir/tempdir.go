package tempdir

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

// Dir ...
type Dir struct {
	Path string
	t    testing.TB
}

// Cleanup ...
func (d Dir) Cleanup() {
	err := os.RemoveAll(d.Path)
	if err != nil {
		d.t.Errorf("tempdir cleanup failed: %v", err)
	}
}

// CheckEmpty ...
// Check whether the given directory is empty. Marks test failed on
// problems, and on any files seen.
func (d Dir) CheckEmpty() {
	f, err := os.Open(d.Path)
	if err != nil {
		d.t.Errorf("Cannot open temp directory: %v", err)
		return
	}
	junk, err := f.Readdirnames(-1)
	if err != nil {
		d.t.Errorf("Cannot list temp directory: %v", err)
		return
	}
	if len(junk) != 0 {
		d.t.Errorf("Temp directory has unexpected junk: %v", junk)
		return
	}
}

// Subdir ...
// Utility function to create a new subdirectory. This can be used to
// have more predictable path names in a test that needs multiple temp
// directories: only make the top one have a random name, name
// subdirectories by their role in the test.
//
// Name must be a valid single path segment, no slashes.
//
// Returns an absolute path.
func (d Dir) Subdir(name string) string {
	p := path.Join(d.Path, name)
	err := os.Mkdir(p, 0700)
	if err != nil {
		d.t.Fatalf("cannot create subdir of temp dir: %v", err)
	}
	return p
}

// New ..
func New(t testing.TB) Dir {
	parent := ""

	// if we are running under "go test", use its temp dir
	arg0 := path.Dir(os.Args[0])
	if path.Base(arg0) == "_test" {
		parent = arg0
	}

	dir, err := ioutil.TempDir(parent, "temp-")
	if err != nil {
		t.Fatalf("cannot create temp directory: %v", err)
	}
	return Dir{Path: dir, t: t}
}
