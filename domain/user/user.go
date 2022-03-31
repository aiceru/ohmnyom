package user

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/aiceru/protonyom/gonyom"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"ohmnyom/internal"
	"ohmnyom/internal/errors"
)

const CtxKeyUid = internal.ContextKey("uid")

const (
	OAuthProviderGoogle = "google"
	OAuthProviderKakao  = "kakao"

	storageSep         = "/"
	storageDirUser     = "users"
	storageDirProfiles = "profiles"
	StorageRoot        = "ohmnyom"
)

type OAuthInfo struct {
	Id    string `firestore:"id,omitempty"`
	Email string `firestore:"email,omitempty"`
}

func (o *OAuthInfo) ToProto() *gonyom.OAuthInfo {
	return &gonyom.OAuthInfo{
		Id:    o.Id,
		Email: o.Email,
	}
}

func OAuthInfoFromProto(info *gonyom.OAuthInfo) *OAuthInfo {
	if info.Id == "" {
		return nil
	}
	return &OAuthInfo{
		Id:    info.Id,
		Email: info.Email,
	}
}

type User struct {
	Id        string                `firestore:"id"`
	Name      string                `firestore:"name,omitempty"`
	Email     string                `firestore:"email,omitempty"`
	Password  string                `firestore:"password,omitempty"`
	OAuthInfo map[string]*OAuthInfo `firestore:"oauthinfo,omitempty"`
	Photourl  string                `firestore:"photourl,omitempty"`
	SignedUp  time.Time             `firestore:"signedup"`
	Pets      []string              `firestore:"pets,omitempty"`
}

var updatableFields = []string{"name", "password", "photourl"}

func IsUpdatableField(field string) bool {
	for _, f := range updatableFields {
		if f == field {
			return true
		}
	}
	return false
}

func NewUser(name, email, hashed string, infos map[string]*OAuthInfo, photourl string) (*User, error) {
	if name == "" || email == "" {
		return nil, errors.NewInvalidParamError("email [%v], name [%v]", email, name)
	}
	if hashed == "" && len(infos) < 1 {
		return nil, errors.NewInvalidParamError("hashed [%v], info [%v]", hashed, infos)
	}
	u := &User{
		Id:        genUid(),
		Name:      name,
		Email:     email,
		Password:  hashed,
		OAuthInfo: infos,
		Photourl:  photourl,
		SignedUp:  time.Now().UTC(),
		Pets:      nil,
	}
	return u, nil
}

func (u *User) HasPet(petId string) bool {
	for _, p := range u.Pets {
		if p == petId {
			return true
		}
	}
	return false
}

func (u *User) ProfileDir() string {
	return strings.Join([]string{storageDirUser, u.Id, storageDirProfiles}, storageSep)
}

func (u *User) NewProfilePath() string {
	timeStr := strconv.FormatInt(time.Now().UTC().UnixNano(), 16)
	return strings.Join([]string{storageDirUser, u.Id, storageDirProfiles, timeStr}, storageSep)
}

func (u *User) ToProto() *gonyom.Account {
	infos := make(map[string]*gonyom.OAuthInfo)
	for provider, info := range u.OAuthInfo {
		infos[provider] = info.ToProto()
	}
	if len(infos) < 1 {
		infos = nil
	}
	return &gonyom.Account{
		Id:          u.Id,
		Name:        u.Name,
		Email:       u.Email,
		HasPassword: u.Password != "",
		Oauthinfo:   infos,
		Photourl:    u.Photourl,
		Signedup:    u.SignedUp.Unix(),
		Pets:        u.Pets,
	}
}

func FromProto(account *gonyom.Account) *User {
	infos := make(map[string]*OAuthInfo)
	for provider, info := range account.Oauthinfo {
		infos[provider] = OAuthInfoFromProto(info)
	}
	if len(infos) < 1 {
		infos = nil
	}
	// except password, signedup
	return &User{
		Id:        account.Id,
		Name:      account.Name,
		Email:     account.Email,
		OAuthInfo: infos,
		Photourl:  account.Photourl,
		Pets:      account.Pets,
	}
}

func genUid() string {
	return xid.New().String()
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.NewInternalError("%v", err)
	}
	return string(hash), nil
}

func CompareHashAndPassword(hashed string, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return errors.NewAuthenticationError("password not mateched")
	}
	return nil
}

type Store interface {
	Get(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByOAuth(ctx context.Context, info *OAuthInfo, provider string) (*User, error)
	Put(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User, path, value string) error
	Delete(ctx context.Context, id string) error
	AddPet(ctx context.Context, id, petId string) error
	DeletePet(ctx context.Context, id, petId string) error
}
