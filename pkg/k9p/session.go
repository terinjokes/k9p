package k9p

import (
	"context"
	"runtime"
	"sync"

	"github.com/docker/go-p9p"
	"go.terinstock.com/k9p/pkg/resources"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type Session struct {
	sync.Mutex
	aname string
	uname string

	client         kubernetes.Interface
	sharedInformer informers.SharedInformerFactory
	refs           map[p9p.Fid]resources.Ref
}

func New(ctx context.Context, client kubernetes.Interface) *Session {
	sharedInformer := informers.NewSharedInformerFactory(client, 0)

	sharedInformer.Core().V1().Namespaces().Informer()
	sharedInformer.Apps().V1().Deployments().Informer()

	go sharedInformer.Start(ctx.Done())
	runtime.Gosched()

	return &Session{
		client:         client,
		sharedInformer: sharedInformer,
		refs:           make(map[p9p.Fid]resources.Ref),
	}
}

func (k *Session) getRef(fid p9p.Fid) (resources.Ref, error) {
	k.Lock()
	defer k.Unlock()

	if fid == p9p.NOFID {
		return nil, p9p.ErrUnknownfid
	}

	ref, found := k.refs[fid]
	if !found {
		return nil, p9p.ErrUnknownfid
	}

	return ref, nil
}

func (k *Session) newRef(fid p9p.Fid, resource resources.Ref) (resources.Ref, error) {
	k.Lock()
	defer k.Unlock()

	if fid == p9p.NOFID {
		return nil, p9p.ErrUnknownfid
	}

	_, ok := k.refs[fid]
	if ok {
		return nil, p9p.ErrDupfid
	}

	ref := resource
	k.refs[fid] = ref
	return ref, nil
}

func (k *Session) Auth(ctx context.Context, afid p9p.Fid, uname string, aname string) (p9p.Qid, error) {
	return p9p.Qid{}, nil
}

func (k *Session) Attach(ctx context.Context, fid p9p.Fid, afid p9p.Fid, uname string, aname string) (p9p.Qid, error) {
	if uname == "" {
		return p9p.Qid{}, p9p.MessageRerror{Ename: "no user"}
	}

	if aname == "" {
		aname = "/"
	}

	k.uname = uname
	k.aname = aname

	ref, err := k.newRef(fid, resources.NewDirRef("/", k, map[string]resources.Ref{
		"namespaces": resources.NewNamespacesRef(k.client, k),
		"cluster":    resources.NewDirRef("cluster", k, map[string]resources.Ref{}),
	}))
	if err != nil {
		return p9p.Qid{}, err
	}

	return ref.Info().Qid, nil
}

func (k *Session) Clunk(ctx context.Context, fid p9p.Fid) error {
	_, err := k.getRef(fid)
	if err != nil {
		return err
	}

	k.Lock()
	defer k.Unlock()
	delete(k.refs, fid)

	return nil
}

func (k *Session) Remove(ctx context.Context, fid p9p.Fid) error {
	return p9p.ErrUnknownMsg
}

func (k *Session) Walk(ctx context.Context, fid p9p.Fid, newfid p9p.Fid, names ...string) ([]p9p.Qid, error) {
	var qids []p9p.Qid

	ref, err := k.getRef(fid)
	if err != nil {
		return qids, err
	}

	current := ref
	for _, name := range names {
		newResource, err := current.Get(name)
		if err != nil {
			break
		}

		qids = append(qids, newResource.Info().Qid)
		current = newResource
	}

	if len(qids) != len(names) {
		return qids, nil
	}

	_, err = k.newRef(newfid, current)
	if err != nil {
		return qids, err
	}

	return qids, nil
}

func (k *Session) Read(ctx context.Context, fid p9p.Fid, p []byte, offset int64) (n int, err error) {
	ref, err := k.getRef(fid)
	if err != nil {
		return 0, err
	}

	return ref.Read(ctx, p, offset)
}

func (k *Session) Write(ctx context.Context, fid p9p.Fid, p []byte, offset int64) (n int, err error) {
	return 0, p9p.ErrUnknownMsg
}

func (k *Session) Open(ctx context.Context, fid p9p.Fid, mode p9p.Flag) (p9p.Qid, uint32, error) {
	ref, err := k.getRef(fid)
	if err != nil {
		return p9p.Qid{}, 0, err
	}

	return ref.Info().Qid, 0, nil
}

func (k *Session) Create(ctx context.Context, parent p9p.Fid, name string, perm uint32, mode p9p.Flag) (p9p.Qid, uint32, error) {
	return p9p.Qid{}, 0, p9p.ErrUnknownMsg
}

func (k *Session) Stat(ctx context.Context, fid p9p.Fid) (p9p.Dir, error) {
	ref, err := k.getRef(fid)
	if err != nil {
		return p9p.Dir{}, err
	}

	return ref.Info(), nil
}

func (k *Session) WStat(ctx context.Context, fid p9p.Fid, dir p9p.Dir) error {
	return p9p.ErrUnknownMsg
}

// Version returns the supported version and msize of the session. This
// can be affected by negotiating or the level of support provided by the
// session implementation.
func (k *Session) Version() (msize int, version string) {
	return p9p.DefaultMSize, p9p.DefaultVersion
}

func (k *Session) WaitForCacheSync(stopCh <-chan struct{}) {
	k.sharedInformer.WaitForCacheSync(stopCh)
}

func (k *Session) GetAuth() (uname, aname string) {
	return k.uname, k.aname
}

func (k *Session) Informer() informers.SharedInformerFactory {
	return k.sharedInformer
}
