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
	is             = "=="
)

type Service struct {
	client *firestore.Client
}

func NewService(ctx context.Context, client *firestore.Client) user.Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Get(ctx context.Context, id string) (*user.User, error) {
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

func (s *Service) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	iter := s.client.Collection(userCollection).Where("email", is, email).Documents(ctx)
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

func (s *Service) GetByOAuth(ctx context.Context, oauthType user.OAuthType, oauthId string) (*user.User, error) {
	iter := s.client.Collection(userCollection).
		Where("oauthid", is, oauthId).
		Where("oauthtype", is, oauthType).
		Documents(ctx)
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
		if u.OAuthType == oauthType && u.OAuthId == oauthId {
			return u, nil
		}
	}
	return nil, errors.NewNotFoundError("User{OAuthType: %v, OAuthId: %v}", oauthType, oauthId)
}

// Put overwrites doc if exists.
func (s *Service) Put(ctx context.Context, user *user.User) error {
	if user == nil {
		return errors.NewInvalidParamError("user: %v", user)
	}
	_, err := s.client.Collection(userCollection).Doc(user.Id).Set(ctx, user)
	if err != nil {
		return errors.New("%v", err)
	}
	return nil
}

// Delete does nothing and returns no error if doc not exists.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.client.Collection(userCollection).Doc(id).Delete(ctx); err != nil {
		return errors.New("%v", err)
	}
	return nil
}
