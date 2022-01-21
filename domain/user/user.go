package user

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"ohmnyom/internal"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/time"
)

type Provider int32

const CtxKeyUid = internal.ContextKey("uid")

const (
	OAuthType_GOOGLE Provider = iota
	OAuthTYpe_KAKAO
)

type OAuthInfo struct {
	Provider Provider `firestore:"provider"`
	Id       string   `firestore:"id,omitempty"`
}

func (o *OAuthInfo) ToProto() *gonyom.OAuthInfo {
	return &gonyom.OAuthInfo{
		Provider: gonyom.OAuthInfo_Provider(o.Provider),
		Id:       o.Id,
	}
}

func OAuthInfoFromProto(info *gonyom.OAuthInfo) *OAuthInfo {
	if info.Id == "" {
		return nil
	}
	return &OAuthInfo{
		Provider: Provider(info.Provider),
		Id:       info.Id,
	}
}

type User struct {
	Id        string       `firestore:"id"`
	Name      string       `firestore:"name,omitempty"`
	Email     string       `firestore:"email,omitempty"`
	Password  string       `firestore:"password,omitempty"`
	OAuthInfo []*OAuthInfo `firestore:"oauthinfo,omitempty"`
	Photourl  string       `firestore:"photourl,omitempty"`
	SignedUp  time.Time    `firestore:"signedup"`
	Pets      []string     `firestore:"pets,omitempty"`
}

func NewUser(name, email, hashed string, info *OAuthInfo, photourl string) (*User, error) {
	if name == "" || email == "" {
		return nil, errors.NewInvalidParamError("email [%v], name [%v]", email, name)
	}
	if hashed == "" && info == nil {
		return nil, errors.NewInvalidParamError("hashed [%v], info [%v]", hashed, info)
	}
	return &User{
		Id:        genUid(),
		Name:      name,
		Email:     email,
		Password:  hashed,
		OAuthInfo: []*OAuthInfo{info},
		Photourl:  photourl,
		SignedUp:  time.Time{},
		Pets:      nil,
	}, nil
}

func (u *User) ToProto() *gonyom.Account {
	infos := make([]*gonyom.OAuthInfo, len(u.OAuthInfo))
	for i, info := range u.OAuthInfo {
		infos[i] = info.ToProto()
	}
	if len(infos) < 1 {
		infos = nil
	}
	return &gonyom.Account{
		Id:        u.Id,
		Name:      u.Name,
		Email:     u.Email,
		Oauthinfo: infos,
		Photourl:  u.Photourl,
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

type Store interface {
	Get(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByOAuth(ctx context.Context, info *OAuthInfo) (*User, error)
	Put(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
