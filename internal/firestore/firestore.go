package firestore

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"ohmnyom/internal/errors"
)

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
