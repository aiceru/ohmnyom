package googleStorage

import (
	"context"
	"io"
	"log"
	"os"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/storage"
)

const RootBucket = "ohmnyom"

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
	wc := s.client.Bucket(object.RootDir).Object(object.Path).NewWriter(ctx)

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

	attrs, err := s.client.Bucket(object.RootDir).Object(object.Path).Attrs(ctx)
	if err != nil {
		return "", nil
	}

	return attrs.MediaLink, nil
}
