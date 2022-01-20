package user

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/internal/errors"
)

type Server struct {
	service Service
	// gonyom.SignApiServer
	gonyom.UnimplementedSignApiServer
}

func NewServer(service Service) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) signUpWithEmailCred(
	ctx context.Context, email, password, name string) (*User, error) {

	_, err := s.service.GetByEmail(ctx, email)
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

	u, err := NewUser(email, name, hashed)
	if err != nil {
		return nil, err
	}

	if err := s.service.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) signUpWithOAuthCred(
	ctx context.Context, email string, cred *gonyom.SignUpRequest_OauthCred, name string, photourl string) (
	*User, error) {

	oauthtype := OAuthType(cred.OauthCred.GetOauthtype())
	oauthid := cred.OauthCred.GetOauthid()

	_, err := s.service.GetByEmail(ctx, email)
	if err == nil { // email exist
		return nil, errors.NewAlreadyExistsError("email [%v]", email)
	} else {
		var notfound *errors.NotFoundError
		if !errors.As(err, &notfound) {
			return nil, err
		}
	}

	_, err = s.service.GetByOAuth(ctx, oauthtype, oauthid)
	if err == nil { // already exists
		return nil, errors.NewAlreadyExistsError("oauthtype [%v], oauthid [%v]")
	} else {
		var notfound *errors.NotFoundError
		if !errors.As(err, &notfound) {
			return nil, err
		}
	}

	u, err := NewOAuthUser(email, oauthtype, oauthid, name, photourl)
	if err != nil {
		return nil, err
	}

	if err := s.service.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) SignUp(ctx context.Context, in *gonyom.SignUpRequest) (*gonyom.SignUpReply, error) {
	name := in.GetName()
	email := in.GetEmail()
	photourl := in.GetPhotourl()

	switch cred := in.GetCredential().(type) {

	case *gonyom.SignUpRequest_Password:
		ret, err := s.signUpWithEmailCred(ctx, email, cred.Password, name)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		return &gonyom.SignUpReply{Account: ret.Proto()}, nil

	case *gonyom.SignUpRequest_OauthCred:
		ret, err := s.signUpWithOAuthCred(ctx, email, cred, name, photourl)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		return &gonyom.SignUpReply{Account: ret.Proto()}, nil

	default:
		return nil, errors.GrpcError(errors.NewUnsupportedError("type %v", cred))
	}
}

func (s *Server) signInWithEmailCred(ctx context.Context, email, password string) (*User, error) {
	u, err := s.service.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := compareHashAndPassword(u.Password, password); err != nil {
		return nil, errors.GrpcError(err)
	}

	return u, nil
}

func (s *Server) signInWithOAuthCred(
	ctx context.Context,
	cred *gonyom.SignInRequest_OauthCred) (*User, error) {
	authtype := OAuthType(cred.OauthCred.GetOauthtype())
	authid := cred.OauthCred.GetOauthid()

	u, err := s.service.GetByOAuth(ctx, authtype, authid)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) SignIn(ctx context.Context, in *gonyom.SignInRequest) (*gonyom.SignInReply, error) {
	email := in.GetEmail()
	switch cred := in.GetCredential().(type) {
	case *gonyom.SignInRequest_Password:
		u, err := s.signInWithEmailCred(ctx, email, cred.Password)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		return &gonyom.SignInReply{Account: u.Proto()}, nil
	case *gonyom.SignInRequest_OauthCred:
		u, err := s.signInWithOAuthCred(ctx, cred)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		return &gonyom.SignInReply{Account: u.Proto()}, nil
	default:
		return nil, errors.GrpcError(errors.NewUnsupportedError("type %v", cred))
	}
}

func (s *Server) SignOut(ctx context.Context, in *gonyom.EmptyParams) (*gonyom.EmptyParams, error) {
	return &gonyom.EmptyParams{}, nil
}
