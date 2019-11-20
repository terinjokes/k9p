package resources

import (
	"github.com/docker/go-p9p"
	"k8s.io/client-go/informers"
)

type Session interface {
	p9p.Session
	GetAuth() (uname, aname string)
	Informer() informers.SharedInformerFactory
}
