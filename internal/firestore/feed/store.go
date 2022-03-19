package feed

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"ohmnyom/domain/feed"
	"ohmnyom/internal/errors"
)

const (
	petCollection  = "pets"
	feedCollection = "feeds"
)

type Store struct {
	client *firestore.Client
}

func New(ctx context.Context, client *firestore.Client) feed.Store {
	return &Store{
		client: client,
	}
}

func (s *Store) Get(ctx context.Context, petId, feedId string) (*feed.Feed, error) {
	doc, err := s.client.Collection(petCollection).Doc(petId).Collection(feedCollection).Doc(feedId).Get(ctx)
	if err != nil {
		return nil, errors.New("%v", err)
	}

	f := &feed.Feed{}
	if err := doc.DataTo(f); err != nil {
		return nil, errors.NewInvalidFormatError("%v", err)
	}

	return f, nil
}

func (s *Store) GetFeedsOfPet(ctx context.Context, petId string, startAfter time.Time, limit int) ([]*feed.Feed, error) {
	p := s.client.Collection(petCollection).Doc(petId)
	page := p.Collection(feedCollection).OrderBy("timestamp", firestore.Desc).
		StartAfter(startAfter).Limit(limit).Documents(ctx)
	docs, err := page.GetAll()
	if err != nil {
		return nil, errors.New("%v", err)
	}

	ret := make([]*feed.Feed, len(docs))
	for i, doc := range docs {
		f := &feed.Feed{}
		if err := doc.DataTo(f); err != nil {
			return nil, errors.NewInvalidFormatError("%v", err)
		}
		ret[i] = f
	}

	return ret, nil
}

func (s *Store) Put(ctx context.Context, feed *feed.Feed) error {
	if feed == nil || feed.Id == "" {
		return errors.NewInvalidParamError("feed: %v", feed)
	}
	_, err := s.client.Collection(petCollection).Doc(feed.PetId).
		Collection(feedCollection).Doc(feed.Id).Create(ctx, feed)
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, petId, feedId string, pathValues map[string]interface{}) error {
	if petId == "" || feedId == "" {
		return errors.NewInvalidParamError("petId: %v, feedId: %v", petId, feedId)
	}

	updates := make([]firestore.Update, len(pathValues))
	i := 0
	for path, value := range pathValues {
		if !feed.IsUpdatableField(path) {
			return errors.NewInvalidParamError("path %v is not updatable", path)
		}
		updates[i].Path = path
		updates[i].Value = value
		i++
	}

	_, err := s.client.Collection(petCollection).Doc(petId).
		Collection(feedCollection).Doc(feedId).Update(ctx, updates)
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, petId, feedId string) error {
	if _, err := s.client.Collection(petCollection).Doc(petId).
		Collection(feedCollection).Doc(feedId).Delete(ctx); err != nil {
		return errors.New("%v", err)
	}
	return nil
}
