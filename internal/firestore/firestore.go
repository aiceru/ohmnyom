package firestore

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"ohmnyom/internal/errors"
)

func NewEmulatorClient(ctx context.Context) *firestore.Client {
	client, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func NewClient(ctx context.Context, projectId, credfile string) (*firestore.Client, error) {
	cred, err := os.ReadFile(credfile)
	if err != nil {
		return nil, errors.New("%v", err)
	}
	client, err := firestore.NewClient(ctx, projectId, option.WithCredentialsJSON(cred))
	if err != nil {
		return nil, errors.New("%v", err)
	}
	return client, nil
}
