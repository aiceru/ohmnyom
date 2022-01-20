package user

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"ohmnyom/internal/errors"
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

func NewUser(email, name, hashed string) (*User, error) {
	if email == "" || hashed == "" || name == "" {
		return nil, errors.NewInvalidParamError("email [%v], hashed [%v], name [%v]", email, hashed, name)
	}
	return &User{
		Id:        genUid(),
		Email:     email,
		Name:      name,
		Password:  hashed,
		Photourl:  "",
		OAuthType: OAuthType_NONE,
		OAuthId:   "",
		SignedUp:  time.Now(),
		Pets:      nil,
	}, nil
}

func NewOAuthUser(email string, oauthtype OAuthType, oauthid, name, photourl string) (*User, error) {
	if email == "" || oauthtype == OAuthType_NONE || oauthid == "" || name == "" {
		return nil, errors.NewInvalidParamError("email [%v], oauthtype [%v], oauthid [%v], name [%v]",
			email, oauthtype, oauthid, name)
	}
	return &User{
		Id:        genUid(),
		Email:     email,
		Name:      name,
		Password:  "",
		Photourl:  photourl,
		OAuthType: oauthtype,
		OAuthId:   oauthid,
		SignedUp:  time.Now(),
		Pets:      nil,
	}, nil
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

func genUid() string {
	return xid.New().String()
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func compareHashAndPassword(hashed string, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return errors.NewAuthenticationError("password not mateched")
	}
	return nil
}

type Service interface {
	Get(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByOAuth(ctx context.Context, oauthType OAuthType, oauthId string) (*User, error)
	Put(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
