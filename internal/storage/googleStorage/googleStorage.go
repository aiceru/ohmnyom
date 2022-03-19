package googleStorage

import (
	"context"
	"io"
	"log"
	"os"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/storage"
)

type Storage struct {
	client *gcs.Client
}

func New(ctx context.Context, credentialJsonPath string) storage.Storage {
	credFile, err := os.Open(credentialJsonPath)
	if err != nil {
		log.Panic(err)
	}
	cred, err := io.ReadAll(credFile)
	if err != nil {
		log.Panic(err)
	}
	client, err := gcs.NewClient(ctx, option.WithCredentialsJSON(cred))
	if err != nil {
		log.Panic(err)
	}
	return &Storage{client: client}
}

func (s *Storage) Upload(ctx context.Context, object *storage.Object) (string, error) {
	wc := s.client.Bucket(object.Root).Object(object.Path).NewWriter(ctx)

	wc.ContentType = object.ContentType
	wc.ACL = []gcs.ACLRule{
		{Entity: gcs.AllUsers, Role: gcs.RoleReader},
		{Entity: "user-aiceru@gmail.com", Role: gcs.RoleOwner},
	}
	if _, err := wc.Write(object.Bytes); err != nil {
		return "", errors.NewInternalError("%v", err)
	}
	if err := wc.Close(); err != nil {
		return "", errors.NewInternalError("%v", err)
	}

	attrs, err := s.client.Bucket(object.Root).Object(object.Path).Attrs(ctx)
	if err != nil {
		return "", nil
	}

	return attrs.MediaLink, nil
}

func (s *Storage) Delete(ctx context.Context, root, path string) error {
	if err := s.client.Bucket(root).Object(path).Delete(ctx); err != nil {
		return errors.NewInternalError("%v", err)
	}
	return nil
}

func (s *Storage) DeleteDir(ctx context.Context, root, dir string) error {
	bucket := s.client.Bucket(root)
	it := bucket.Objects(ctx, &gcs.Query{
		Prefix: dir,
	})
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return errors.NewInternalError("%v", err)
		}
		if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
			return errors.NewInternalError("%v", err)
		}
	}
	return nil
}
