package user

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/jwt"
)

type Server struct {
	userStore Store
	*jwt.Manager
	gonyom.UnimplementedSignApiServer
}

func NewServer(service Store) *Server {
	return &Server{
		userStore: service,
	}
}

func (s *Server) signUpWithEmail(
	ctx context.Context, name, email, password, photourl string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.NewInvalidParamError("name [%v], email [%v], password [%v]",
			name, email, password)
	}

	_, err := s.userStore.GetByEmail(ctx, email)
	if err == nil { // email exist
		return nil, errors.NewAlreadyExistsError("email [%v]", email)
	} else {
		var notfound *errors.NotFoundError
		if !errors.As(err, &notfound) {
			return nil, err
		}
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	u, err := NewUser(name, email, hashed, nil, photourl)
	if err != nil {
		return nil, err
	}

	if err := s.userStore.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) signUpWithOAuthInfo(
	ctx context.Context, name, email string, info *OAuthInfo, photourl string) (
	*User, error) {
	if name == "" || email == "" || info == nil {
		return nil, errors.NewInvalidParamError("name [%v], email [%v], info [%v]")
	}

	_, err := s.userStore.GetByEmail(ctx, email)
	if err == nil { // email exist
		return nil, errors.NewAlreadyExistsError("email [%v]", email)
	} else {
		var notfound *errors.NotFoundError
		if !errors.As(err, &notfound) {
			return nil, err
		}
	}

	_, err = s.userStore.GetByOAuth(ctx, info)
	if err == nil { // already exists
		return nil, errors.NewAlreadyExistsError("info [%v]", info)
	} else {
		var notfound *errors.NotFoundError
		if !errors.As(err, &notfound) {
			return nil, err
		}
	}

	u, err := NewUser(name, email, "", info, photourl)
	if err != nil {
		return nil, err
	}

	if err := s.userStore.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) SignUp(ctx context.Context, in *gonyom.SignUpRequest) (*gonyom.SignUpReply, error) {
	var u *User
	var err error

	name := in.GetName()
	email := in.GetEmail()
	photourl := in.GetPhotourl()

	switch cred := in.GetCredential().(type) {
	case *gonyom.SignUpRequest_Password:
		u, err = s.signUpWithEmail(ctx, name, email, cred.Password, photourl)
	case *gonyom.SignUpRequest_Oauthinfo:
		u, err = s.signUpWithOAuthInfo(ctx, name, email, OAuthInfoFromProto(cred.Oauthinfo), photourl)
	default:
		err = errors.NewUnsupportedError("type %v", cred)
	}
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	token, err := s.NewAuthToken(u.Id)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.SignUpReply{Account: u.ToProto(), Token: token}, nil
}

func (s *Server) signInWithEmail(ctx context.Context, email, password string) (*User, error) {
	u, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := compareHashAndPassword(u.Password, password); err != nil {
		return nil, errors.GrpcError(err)
	}

	return u, nil
}

func (s *Server) signInWithOAuthInfo(ctx context.Context, info *OAuthInfo) (*User, error) {
	u, err := s.userStore.GetByOAuth(ctx, info)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) SignIn(ctx context.Context, in *gonyom.SignInRequest) (*gonyom.SignInReply, error) {
	var u *User
	var err error
	email := in.GetEmail()

	switch cred := in.GetCredential().(type) {
	case *gonyom.SignInRequest_Password:
		u, err = s.signInWithEmail(ctx, email, cred.Password)
	case *gonyom.SignInRequest_Oauthinfo:
		u, err = s.signInWithOAuthInfo(ctx, OAuthInfoFromProto(cred.Oauthinfo))
	default:
		err = errors.NewUnsupportedError("type %v", cred)
	}
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	token, err := s.NewAuthToken(u.Id)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.SignInReply{Account: u.ToProto(), Token: token}, nil
}

func (s *Server) SignOut(ctx context.Context, in *gonyom.EmptyParams) (*gonyom.EmptyParams, error) {
	return &gonyom.EmptyParams{}, nil
}
