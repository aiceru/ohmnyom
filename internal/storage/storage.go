package storage

import "context"

type Object struct {
	RootDir     string
	Path        string
	ContentType string
	Bytes       []byte
}

type Storage interface {
	Upload(ctx context.Context, object *Object) (string, error)
}
