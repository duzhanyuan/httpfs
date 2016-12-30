package fsapi

import (
	"io"
	//"log"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var _ fs.Node = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)
var _ fs.NodeAccesser = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)
var _ fs.HandleWriter = (*File)(nil)
var _ fs.HandleReleaser = (*File)(nil)

// File ...
type File struct {
	sync.RWMutex
	attr    fuse.Attr
	path    string
	created bool
	fs      *HTTPFS
	handle  *Handle
}

// Access ...
func (f *File) Access(ctx context.Context, req *fuse.AccessRequest) error {
	//log.Printf("file.Access(%s)\n", f.path)

	//log.Printf(" ctx=+%v\n", ctx)
	//log.Printf(" req=+%v\n", req)

	return nil
}

// Attr ...
func (f *File) Attr(ctx context.Context, o *fuse.Attr) error {
	//log.Printf("file.Attr(%s)\n", f.path)

	f.RLock()
	err := f.readAttr()
	if err != nil {
		//log.Printf(" E: %s\n", err)
	}

	*o = f.attr

	//log.Printf(" attr=%s\n", f.attr)
	//log.Printf(" mtime=%d\n", f.attr.Mtime.Unix())

	f.RUnlock()
	return nil
}

func (f *File) readAttr() error {
	stats, err := f.fs.client.Stat(f.path)
	if err != nil {
		return err
	}

	f.attr.Size = uint64(stats.Size())
	f.attr.Mtime = stats.ModTime()
	f.attr.Mode = stats.Mode()

	return nil
}

// Open ...
func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	//log.Printf("file.Open(%s, %d, %d)\n", f.path, int(req.Flags), f.attr.Mode)

	//log.Printf(" req=%s\n", req)

	handle := Handle{
		f:     f,
		path:  f.path,
		flags: int(req.Flags),
		perm:  f.attr.Mode,

		client: f.fs.client,
	}

	f.handle = &handle

	return f, nil
}

// Release ...
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	//log.Printf("file.Release(%s)\n", f.path)
	return f.handle.Close()
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	//log.Printf("file.Read(%s)\n", f.path)

	f.RLock()
	defer f.RUnlock()

	//log.Printf(" req=%s\n", req)

	if f.handle == nil {
		//log.Println(" E: file not open")
		return fuse.ENOTSUP
	}

	resp.Data = resp.Data[:req.Size]
	n, err := f.handle.ReadAt(resp.Data, req.Offset)
	if err != nil && err != io.EOF {
		//log.Printf(" E: %s\n", err)
		return err
	}
	resp.Data = resp.Data[:n]
	//log.Printf(" %d bytes read\n", n)

	return nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	//log.Printf("file.Write(%s, %q)\n", f.path, req.Data)

	f.Lock()
	defer f.Unlock()

	//log.Printf(" req=%s\n", req)

	if f.handle == nil {
		//log.Println(" E: file not open")
		return fuse.ENOTSUP
	}

	n, err := f.handle.WriteAt(req.Data, int(req.FileFlags), req.Offset)
	if err != nil {
		//log.Printf(" E: %s\n", err)
		return err
	}
	resp.Size = n
	//log.Printf(" %d bytes written\n", n)

	return nil
}

var _ fs.NodeSetattrer = (*File)(nil)

// Setattr ...
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	//log.Printf("file.Setattr(%s)\n", f.path)

	f.Lock()
	defer f.Unlock()

	//log.Printf(" req=%q\n", req)

	// f.dirty = dirty

	valid := req.Valid

	if valid.Size() {
		err := f.fs.client.Truncate(f.path, req.Size)
		if err != nil {
			//log.Printf(" E: %s\n", err)
			return err
		}
		valid &^= fuse.SetattrSize
	}

	if valid.Mode() {
		err := f.fs.client.Chmod(f.path, req.Mode)
		if err != nil {
			//log.Printf(" E: %s\n", err)
			return err
		}
		valid &^= fuse.SetattrMode
	}

	// things we don't need to explicitly handle
	valid &^= fuse.SetattrLockOwner | fuse.SetattrHandle

	if valid != 0 {
		// don't let an unhandled operation slip by without error
		//log.Printf(" E: Setattr did not handle %v\n", valid)
		return fuse.ENOSYS
	}
	return nil
}
