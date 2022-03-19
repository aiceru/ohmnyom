package pet

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ohmnyom/domain/pet"
	"ohmnyom/internal/errors"
)

const (
	petCollection = "pets"
	operatorIn    = "in"
)

type Store struct {
	client *firestore.Client
}

func New(ctx context.Context, client *firestore.Client) pet.Store {
	return &Store{
		client: client,
	}
}

func (s *Store) Get(ctx context.Context, id string) (*pet.Pet, error) {
	snapshot, err := s.client.Collection(petCollection).Doc(id).Get(ctx)
	switch status.Code(err) {
	case codes.OK:
		p := &pet.Pet{}
		if suberr := snapshot.DataTo(p); suberr != nil {
			return nil, errors.NewInvalidFormatError("%v", suberr)
		}
		return p, nil
	case codes.NotFound:
		return nil, errors.NewNotFoundError("Pet{Id: %v}", id)
	}
	return nil, errors.New("%v", err)
}

func (s *Store) GetList(ctx context.Context, ids []string) (pet.List, error) {
	query := s.client.Collection(petCollection).Where("id", operatorIn, ids).Documents(ctx)
	ret := make([]*pet.Pet, 0)
	for {
		doc, err := query.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		p := &pet.Pet{}
		if suberr := doc.DataTo(p); suberr != nil {
			return nil, errors.NewInvalidFormatError("%v", suberr)
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func (s *Store) Put(ctx context.Context, p *pet.Pet) error {
	if p == nil {
		return errors.NewInvalidParamError("p: %v", p)
	}
	_, err := s.client.Collection(petCollection).Doc(p.Id).Create(ctx, p)
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, id string, pathValues map[string]interface{}) error {
	if id == "" {
		return errors.NewInvalidParamError("id: %v", id)
	}

	updates := make([]firestore.Update, len(pathValues))
	i := 0
	for path, value := range pathValues {
		if !pet.IsUpdatableField(path) {
			return errors.NewInvalidParamError("path %v is not updatable", path)
		}
		updates[i].Path = path
		updates[i].Value = value
		i++
	}

	_, err := s.client.Collection(petCollection).Doc(id).Update(ctx, updates)
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if _, err := s.client.Collection(petCollection).Doc(id).Delete(ctx); err != nil {
		return errors.New("%v", err)
	}
	return nil
}
