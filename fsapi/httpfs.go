package fsapi

import (
	"os"
	"sync/atomic"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// HTTPFS ...
type HTTPFS struct {
	root   *Dir
	NodeID uint64
	size   int64

	client *Client
}

// Compile-time interface checks.
var _ fs.FS = (*HTTPFS)(nil)
var _ fs.FSStatfser = (*HTTPFS)(nil)

// DefaultFileMode ...
const DefaultFileMode = os.FileMode(int(0777))

// NewHTTPFS ...
func NewHTTPFS(url string, tlsverify bool) *HTTPFS {
	fs := &HTTPFS{
		client: NewClient(url, tlsverify),
	}
	fs.root = fs.newDir("/", os.ModeDir|DefaultFileMode)
	if fs.root.attr.Inode != 1 {
		panic("Root node should have been assigned id 1")
	}
	return fs
}

func (m *HTTPFS) nextID() uint64 {
	return atomic.AddUint64(&m.NodeID, 1)
}

func (m *HTTPFS) newDir(path string, mode os.FileMode) *Dir {
	n := time.Now()
	return &Dir{
		attr: fuse.Attr{
			Inode:  m.nextID(),
			Atime:  n,
			Mtime:  n,
			Ctime:  n,
			Crtime: n,
			Mode:   os.ModeDir | mode,
		},
		path: path,
		fs:   m,
	}
}

func (m *HTTPFS) newFile(path string, mode os.FileMode) *File {
	n := time.Now()
	return &File{
		attr: fuse.Attr{
			Inode:  m.nextID(),
			Atime:  n,
			Mtime:  n,
			Ctime:  n,
			Crtime: n,
			Mode:   mode,
		},
		path: path,
		fs:   m,
	}
}

// Root ...
func (m *HTTPFS) Root() (fs.Node, error) {
	return m.root, nil
}

// Statfs ...
func (m *HTTPFS) Statfs(ctx context.Context, req *fuse.StatfsRequest, res *fuse.StatfsResponse) error {
	/*
		// TODO: Build an API for this in httpfs
		s := syscall.Statfs_t{}

		err := syscall.Statfs(f.path, &s)
		if err != nil {
			//log.Println("DRIVE | Statfs syscall failed; ", err)
			return err
		}

		res.Blocks = s.Blocks
		res.Bfree = s.Bfree
		res.Bavail = s.Bavail
		res.Ffree = s.Ffree
		res.Bsize = s.Bsize
	*/

	return nil
}
