package user

import (
	"context"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	local "ohmnyom/internal/firestore"
)

const firestoreEmulatorHostStr = "FIRESTORE_EMULATOR_HOST"

func setupTestData(ctx context.Context, client *firestore.Client) {
	for _, user := range users {
		_, err := client.Collection(userCollection).Doc(user.Id).Set(ctx, user)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func teardownTestData(ctx context.Context, client *firestore.Client) {
	iter := client.Collection(userCollection).Limit(10).Documents(ctx)
	batch := client.Batch()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		batch.Delete(doc.Ref)
	}
	_, err := batch.Commit(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	local.KillEmulator()
	local.RunEmulator()

	var result int

	defer func() {
		local.KillEmulator()
		os.Exit(result)
	}()

	result = m.Run()
}

func newTestClient(ctx context.Context) *firestore.Client {
	client, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}
	return client
}
