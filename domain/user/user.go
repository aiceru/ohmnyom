package user

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/internal/time"
)

type OAuthType int32

const (
	OAuthType_NONE OAuthType = iota
	OAuthType_GOOGLE
	OAuthTYpe_KAKAO
)

type User struct {
	Id        string    `firestore:"id"`
	Email     string    `firestore:"email,omitempty"`
	Name      string    `firestore:"name,omitempty"`
	Password  string    `firestore:"password,omitempty"`
	Photourl  string    `firestore:"photourl,omitempty"`
	OAuthType OAuthType `firestore:"oauthtype"`
	OAuthId   string    `firestore:"oauthid,omitempty"`
	SignedUp  time.Time `firestore:"signedup"`
	Pets      []string  `firestore:"pets,omitempty"`
}

func (u *User) Proto() *gonyom.Account {
	return &gonyom.Account{
		Id:        u.Id,
		Email:     u.Email,
		Name:      u.Name,
		Photourl:  u.Photourl,
		Oauthtype: gonyom.OAuthType(u.OAuthType),
		Oauthid:   u.OAuthId,
		Signedup:  u.SignedUp.Proto(),
		Pets:      u.Pets,
	}
}

type Service interface {
	Get(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) ([]*User, error)
	GetByOAuth(ctx context.Context, oauthType OAuthType, oauthId string) (*User, error)
	Put(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
