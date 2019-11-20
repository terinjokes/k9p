package resources

import (
	"context"
	"io"
	"math/rand"
	"strconv"
	"time"

	"github.com/docker/go-p9p"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	"sigs.k8s.io/yaml"
)

type Deployments struct {
	namespace          string
	deploymentInformer appsv1.DeploymentInformer
	session            Session
	info               *p9p.Dir
	readdir            *p9p.Readdir
}

func NewDeployments(namespace string, session Session) *Deployments {
	deploymentInformer := session.Informer().Apps().V1().Deployments()
	return &Deployments{
		namespace:          namespace,
		deploymentInformer: deploymentInformer,
		session:            session,
	}
}

func (r *Deployments) Info() p9p.Dir {
	if r.info != nil {
		return *r.info
	}

	dir := p9p.Dir{}
	dir.Qid.Path = rand.Uint64()
	dir.Qid.Version = 0

	dir.Name = "deployments"
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

func (r *Deployments) Get(name string) (Ref, error) {
	deployment, err := r.deploymentInformer.Lister().Deployments(r.namespace).Get(name)
	if apierrors.IsNotFound(err) {
		return nil, p9p.ErrNotfound
	}
	if err != nil {
		return nil, err
	}

	return NewDeploymentRef(deployment, r.session), nil
}

func (r *Deployments) Read(ctx context.Context, p []byte, offset int64) (n int, err error) {
	if r.readdir != nil {
		return r.readdir.Read(ctx, p, offset)
	}

	deployments, err := r.deploymentInformer.Lister().Deployments(r.namespace).List(labels.Everything())
	if err != nil {
		return 0, err
	}

	deploymentRefs := make([]Ref, 0, len(deployments))

	for _, deployment := range deployments {
		deployment := deployment
		deploymentRefs = append(deploymentRefs, NewDeploymentRef(deployment, r.session))
	}

	r.readdir = p9p.NewReaddir(p9p.NewCodec(), func() (p9p.Dir, error) {
		if len(deploymentRefs) == 0 {
			return p9p.Dir{}, io.EOF
		}

		deployment := deploymentRefs[0]
		deploymentRefs = deploymentRefs[1:]

		return deployment.Info(), nil
	})

	n, err = r.readdir.Read(ctx, p, offset)

	return n, err
}

type DeploymentRef struct {
	deployment *v1.Deployment
	session    Session
	info       *p9p.Dir
	readdir    *p9p.Readdir
	children   map[string]Ref
}

func NewDeploymentRef(deployment *v1.Deployment, session Session) *DeploymentRef {
	y, _ := yaml.Marshal(deployment)
	children := map[string]Ref{
		"data.yaml": &Static{
			name:    "data.yaml",
			content: y,
			session: session,
		},
		"scale": &Static{
			name:    "scale",
			content: []byte(strconv.Itoa(int(*deployment.Spec.Replicas))),
			session: session,
		},
	}
	return &DeploymentRef{
		deployment: deployment,
		session:    session,
		children:   children,
	}
}

func (r *DeploymentRef) Info() p9p.Dir {
	if r.info != nil {
		return *r.info
	}

	dir := p9p.Dir{}
	dir.Qid.Path = rand.Uint64()
	dir.Qid.Version = 0

	dir.Name = r.deployment.Name
	dir.Mode = 0664
	dir.Length = 0
	dir.AccessTime = r.deployment.CreationTimestamp.Time
	dir.ModTime = r.deployment.CreationTimestamp.Time
	dir.MUID = "none"

	uname, _ := r.session.GetAuth()
	dir.UID = uname
	dir.GID = uname

	dir.Qid.Type |= p9p.QTDIR
	dir.Mode |= p9p.DMDIR
	r.info = &dir

	return dir
}

func (r *DeploymentRef) Get(name string) (Ref, error) {
	ref, ok := r.children[name]
	if !ok {
		return nil, p9p.ErrNotfound
	}

	return ref, nil
}

func (r *DeploymentRef) Read(ctx context.Context, p []byte, offset int64) (int, error) {
	if r.readdir != nil {
		return r.readdir.Read(ctx, p, offset)
	}

	dir := make([]p9p.Dir, 0, len(r.children))
	for _, child := range r.children {
		dir = append(dir, child.Info())
	}

	r.readdir = p9p.NewFixedReaddir(p9p.NewCodec(), dir)
	return r.readdir.Read(ctx, p, offset)
}
