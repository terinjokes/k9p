package resources

import (
	"context"
	"math/rand"
	"time"

	"github.com/docker/go-p9p"
)

type DirRef struct {
	path     string
	info     p9p.Dir
	session  Session
	children map[string]Ref
	readdir  *p9p.Readdir
}

func NewDirRef(path string, session Session, children map[string]Ref) *DirRef {
	d := &DirRef{
		path:     path,
		session:  session,
		children: children,
	}
	d.info = d.createInfo()

	return d
}

func (d *DirRef) createInfo() p9p.Dir {
	dir := p9p.Dir{}
	dir.Qid.Path = rand.Uint64()
	dir.Qid.Version = 0

	dir.Name = d.path
	dir.Mode = 0664
	dir.Length = 0
	dir.AccessTime = time.Now()
	dir.ModTime = time.Now()
	dir.MUID = "none"

	uname, _ := d.session.GetAuth()
	dir.UID = uname
	dir.GID = uname

	dir.Qid.Type |= p9p.QTDIR
	dir.Mode |= p9p.DMDIR

	return dir
}

func (d *DirRef) Info() p9p.Dir {
	return d.info
}

func (d *DirRef) Get(name string) (Ref, error) {
	child, ok := d.children[name]
	if !ok {
		return nil, p9p.ErrNotfound
	}

	return child, nil
}

func (d *DirRef) Read(ctx context.Context, p []byte, offset int64) (n int, err error) {
	if d.readdir != nil {
		return d.readdir.Read(ctx, p, offset)
	}

	dir := make([]p9p.Dir, 0, len(d.children))
	for _, child := range d.children {
		dir = append(dir, child.Info())
	}
	d.readdir = p9p.NewFixedReaddir(p9p.NewCodec(), dir)
	return d.readdir.Read(ctx, p, offset)
}
