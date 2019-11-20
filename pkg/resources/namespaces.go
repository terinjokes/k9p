package resources

import (
	"context"
	"io"
	"math/rand"
	"time"

	"github.com/docker/go-p9p"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespacesRef struct {
	client            kubernetes.Interface
	namespaceInformer corev1.NamespaceInformer
	session           Session
	info              *p9p.Dir
	readdir           *p9p.Readdir
}

func NewNamespacesRef(client kubernetes.Interface, session Session) *NamespacesRef {
	namespaceInformer := session.Informer().Core().V1().Namespaces()

	return &NamespacesRef{
		client:            client,
		namespaceInformer: namespaceInformer,
		session:           session,
	}
}

func (r *NamespacesRef) Info() p9p.Dir {
	if r.info != nil {
		return *r.info
	}

	dir := p9p.Dir{}
	dir.Qid.Path = rand.Uint64()
	dir.Qid.Version = 0

	dir.Name = "namespaces"
	dir.Mode = 0664
	dir.Length = 0
	dir.AccessTime = time.Now()
	dir.ModTime = time.Now()
	dir.MUID = "none"

	uname, _ := r.session.GetAuth()
	dir.UID = uname
	dir.GID = uname

	dir.Qid.Type |= p9p.QTDIR
	dir.Mode |= p9p.DMDIR
	r.info = &dir

	return dir
}

func (r *NamespacesRef) Get(name string) (Ref, error) {
	namespace, err := r.namespaceInformer.Lister().Get(name)
	if apierrors.IsNotFound(err) {
		return nil, p9p.ErrNotfound
	}
	if err != nil {
		return nil, err
	}

	return &NamespaceRef{
		namespace: namespace,
		client:    r.client,
		session:   r.session,
	}, nil
}

func (r *NamespacesRef) Read(ctx context.Context, p []byte, offset int64) (n int, err error) {
	if r.readdir != nil {
		return r.readdir.Read(ctx, p, offset)
	}

	namespaces, err := r.namespaceInformer.Lister().List(labels.Everything())
	if err != nil {
		return 0, err
	}

	namespaceRefs := make([]NamespaceRef, 0, len(namespaces))

	for _, namespace := range namespaces {
		namespace := namespace
		namespaceRefs = append(namespaceRefs, NamespaceRef{
			namespace: namespace,
			client:    r.client,
			session:   r.session,
		})
	}

	r.readdir = p9p.NewReaddir(p9p.NewCodec(), func() (p9p.Dir, error) {
		if len(namespaceRefs) == 0 {
			return p9p.Dir{}, io.EOF
		}

		ns := namespaceRefs[0]
		namespaceRefs = namespaceRefs[1:]

		return ns.Info(), nil
	})

	return r.readdir.Read(ctx, p, offset)
}
