package utils_test

import (
	"os"
	"path"
	"testing"

	"github.com/prologic/httpfs/utils"
	"github.com/prologic/httpfs/utils/tempdir"

	"github.com/stretchr/testify/assert"
)

func TestReadDirFile(t *testing.T) {
	assert := assert.New(t)

	tmp := tempdir.New(t)
	defer tmp.Cleanup()

	f, err := os.Create(path.Join(tmp.Path, "hello.txt"))
	assert.Nil(err)

	f.WriteString("Hello World!")
	f.Close()

	items, err := utils.ReadDir(tmp.Path)
	assert.Nil(err)

	item := items[0]

	assert.Equal(item.Name, "hello.txt")
	assert.EqualValues(item.Size, 12)
	assert.NotZero(item.Mode, 0)
	assert.False(item.IsDir)
	assert.NotZero(item.ModTime)
}

var booltests = []struct {
	in  string
	out bool
	def bool
}{
	{"yes", true, true},
	{"Yes", true, true},
	{"y", true, true},
	{"Y", true, true},
	{"1", true, true},
	{"yeah", true, true},
	{"no", false, false},
	{"No", false, false},
	{"n", false, false},
	{"N", false, false},
	{"0", false, false},
	{"nah", false, false},
}

func TestSafeParseBool(t *testing.T) {
	for _, tt := range booltests {
		b := utils.SafeParseBool(tt.in, tt.def)
		assert.Equal(t, b, tt.out)
	}
}

var inttests = []struct {
	in  string
	out int
	def int
}{
	{"1", 1, 1},
	{"1234", 1234, 1},
	{"0", 0, 0},
	{"asdf", 0, 0},
	{"4294967297", 0, 0},
}

func TestSafeParseInt(t *testing.T) {
	for _, tt := range inttests {
		i := utils.SafeParseInt(tt.in, tt.def)
		assert.Equal(t, i, tt.out)
	}
}

var int64tests = []struct {
	in  string
	out int64
	def int64
}{
	{"1", 1, 1},
	{"1234", 1234, 1},
	{"0", 0, 0},
	{"asdf", 0, 0},
	{"4294967297", 4294967297, 0},
}

func TestSafeParseInt64(t *testing.T) {
	for _, tt := range int64tests {
		i := utils.SafeParseInt64(tt.in, tt.def)
		assert.Equal(t, i, tt.out)
	}
}

func TestSafeStatSize(t *testing.T) {
	assert := assert.New(t)

	tmp := tempdir.New(t)
	defer tmp.Cleanup()

	f, err := os.Create(path.Join(tmp.Path, "hello.txt"))
	assert.Nil(err)

	f.WriteString("Hello World!")
	f.Close()

	size := utils.SafeStatSize(path.Join(tmp.Path, "hello.txt"))
	assert.EqualValues(size, 12)
}
