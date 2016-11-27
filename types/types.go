package types

// Entry ...
type Entry struct {
	Name    string
	Size    int64
	Mode    uint32
	ModTime int64
	IsDir   bool
}
