package user

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ohmnyom/domain/user"
	"ohmnyom/internal/errors"
)

const (
	userCollection = "users"
	operatorIs     = "=="
)

type Store struct {
	client *firestore.Client
}

func New(ctx context.Context, client *firestore.Client) user.Store {
	return &Store{
		client: client,
	}
}

func (s *Store) Get(ctx context.Context, id string) (*user.User, error) {
	snapshot, err := s.client.Collection(userCollection).Doc(id).Get(ctx)
	if status.Code(err) == codes.OK {
		u := &user.User{}
		if suberr := snapshot.DataTo(u); suberr != nil {
			return nil, errors.NewInvalidFormatError("%v", suberr)
		}
		return u, nil
	} else if status.Code(err) == codes.NotFound {
		return nil, errors.NewNotFoundError("User{Id: %v}", id)
	} else {
		return nil, errors.New("%v", err)
	}
}

func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	iter := s.client.Collection(userCollection).Where("email", operatorIs, email).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.New("%v", err)
		}
		u := &user.User{}
		if suberr := doc.DataTo(u); suberr != nil {
			// best effort
			return nil, errors.NewInvalidFormatError("%v", suberr)
		}
		return u, nil
	}
	return nil, errors.NewNotFoundError("User{Email: %v}", email)
}

func (s *Store) GetByOAuth(ctx context.Context, info *user.OAuthInfo, provider string) (*user.User, error) {
	iter := s.client.Collection(userCollection).
		Where("oauthinfo."+provider, operatorIs, info).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.New("%v", err)
		}
		u := &user.User{}
		if suberr := doc.DataTo(u); suberr != nil {
			return nil, errors.NewInvalidFormatError("%v", suberr)
		}
		return u, nil
	}
	return nil, errors.NewNotFoundError("User{OAuthInfo: %v}", info)
}

func (s *Store) Put(ctx context.Context, user *user.User) error {
	if user == nil {
		return errors.NewInvalidParamError("user: %v", user)
	}
	_, err := s.client.Collection(userCollection).Doc(user.Id).Create(ctx, user)
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, u *user.User, path, value string) error {
	if u == nil {
		return errors.NewInvalidParamError("u: %v", u)
	}

	if !user.IsUpdatableField(path) {
		return errors.NewInvalidParamError("path %v is not updatable", path)
	}

	_, err := s.client.Collection(userCollection).Doc(u.Id).Update(ctx, []firestore.Update{
		{Path: path, Value: value},
	})
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

// Delete does nothing and returns no error if doc not exists.
func (s *Store) Delete(ctx context.Context, id string) error {
	if _, err := s.client.Collection(userCollection).Doc(id).Delete(ctx); err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) AddPet(ctx context.Context, id, petId string) error {
	if id == "" || petId == "" {
		return errors.NewInvalidParamError("id: %v, petId: %v", id, petId)
	}
	_, err := s.client.Collection(userCollection).Doc(id).Update(ctx,
		[]firestore.Update{
			{Path: "pets", Value: firestore.ArrayUnion(petId)},
		})
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

func (s *Store) DeletePet(ctx context.Context, id, petId string) error {
	if id == "" || petId == "" {
		return errors.NewInvalidParamError("id: %v, petId: %v", id, petId)
	}
	_, err := s.client.Collection(userCollection).Doc(id).Update(ctx,
		[]firestore.Update{
			{Path: "pets", Value: firestore.ArrayRemove(petId)},
		})
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}
