package resources

import (
	"context"

	"github.com/docker/go-p9p"
)

type Ref interface {
	Info() p9p.Dir
	Get(name string) (Ref, error)
	Read(ctx context.Context, p []byte, offset int64) (n int, err error)
}
