package resources

import (
	"context"
	"math/rand"
	"time"

	"github.com/docker/go-p9p"
	"github.com/rs/zerolog/log"
)

type Static struct {
	name    string
	offset  int64
	content []byte
	info    *p9p.Dir
	session Session
}

func (r *Static) Info() p9p.Dir {
	if r.info != nil {
		return *r.info
	}

	dir := p9p.Dir{}
	dir.Qid.Path = rand.Uint64()
	dir.Qid.Version = 0

	dir.Name = r.name
	dir.Mode = 0664
	dir.Length = 0
	dir.AccessTime = time.Now()
	dir.ModTime = time.Now()
	dir.MUID = "none"

	uname, _ := r.session.GetAuth()
	dir.UID = uname
	dir.GID = uname

	dir.Qid.Type |= p9p.QTFILE

	r.info = &dir
	return dir
}

func (r *Static) Get(name string) (Ref, error) {
	return nil, p9p.ErrWalknodir
}

func (r *Static) Read(ctx context.Context, p []byte, offset int64) (n int, err error) {
	log.Debug().Int64("offset", r.offset).Str("name", r.name).Send()
	if offset != r.offset {
		return 0, p9p.ErrBadoffset
	}

	n = copy(p, r.content[offset:])
	r.offset += int64(n)

	return n, nil
}
