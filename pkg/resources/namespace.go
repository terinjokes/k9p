package resources

import (
	"context"
	"math/rand"

	"github.com/docker/go-p9p"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespaceRef struct {
	namespace *v1.Namespace
	client    kubernetes.Interface
	session   Session
	info      *p9p.Dir
	readdir   *p9p.Readdir
}

func (r *NamespaceRef) Info() p9p.Dir {
	if r.info != nil {
		return *r.info
	}

	dir := p9p.Dir{}
	dir.Qid.Path = rand.Uint64()
	dir.Qid.Version = uint32(r.namespace.Generation)

	dir.Name = r.namespace.Name
	dir.Mode = 0664
	dir.Length = 0
	dir.AccessTime = r.namespace.CreationTimestamp.Time
	dir.ModTime = r.namespace.CreationTimestamp.Time
	dir.MUID = "none"

	uname, _ := r.session.GetAuth()
	dir.UID = uname
	dir.GID = uname

	dir.Qid.Type |= p9p.QTDIR
	dir.Mode |= p9p.DMDIR

	r.info = &dir
	return dir
}

func (r *NamespaceRef) Get(name string) (Ref, error) {
	switch name {
	case "deployments":
		return NewDeployments(r.namespace.Name, r.session), nil
	}

	return nil, p9p.ErrNotfound
}

func (r *NamespaceRef) Read(ctx context.Context, p []byte, offset int64) (n int, err error) {
	if r.readdir != nil {
		return r.readdir.Read(ctx, p, offset)
	}

	deployments := NewDeployments(r.namespace.Name, r.session)
	dir := []p9p.Dir{
		deployments.Info(),
	}

	r.readdir = p9p.NewFixedReaddir(p9p.NewCodec(), dir)
	return r.readdir.Read(ctx, p, offset)
}
