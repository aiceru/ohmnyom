package storage

import "context"

type Object struct {
	Root        string
	Path        string
	ContentType string
	Bytes       []byte
}

type Storage interface {
	Upload(ctx context.Context, object *Object) (string, error)
	Delete(ctx context.Context, root, path string) error
	DeleteDir(ctx context.Context, root, dir string) error
}
