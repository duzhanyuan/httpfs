package httpfs_test

import (
	"testing"

	"github.com/prologic/httpfs/utils/tempdir"
)

// TestSimple ...
func TestSimple(t *testing.T) {
	tmp := tempdir.New(t)
	defer tmp.Cleanup()

}
