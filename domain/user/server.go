package user

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/jwt"
)

type Server struct {
	userStore  Store
	jwtManager *jwt.Manager
	gonyom.UnimplementedSignApiServer
	gonyom.UnimplementedAccountApiServer
}

func NewServer(store Store, jwtManager *jwt.Manager) *Server {
	return &Server{
		userStore:  store,
		jwtManager: jwtManager,
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
	ctx context.Context, name, email string, info *OAuthInfo, provider, photourl string) (
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

	_, err = s.userStore.GetByOAuth(ctx, info, provider)
	if err == nil { // already exists
		return nil, errors.NewAlreadyExistsError("info [%v]", info)
	} else {
		var notfound *errors.NotFoundError
		if !errors.As(err, &notfound) {
			return nil, err
		}
	}

	u, err := NewUser(name, email, "", map[string]*OAuthInfo{provider: info}, photourl)
	if err != nil {
		return nil, err
	}

	if err := s.userStore.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) signInWithEmail(ctx context.Context, email, password string) (*User, error) {
	u, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if u.Password == "" {
		return nil, errors.NewAuthenticationError("password not set")
	}

	if err := compareHashAndPassword(u.Password, password); err != nil {
		return nil, errors.NewAuthenticationError("password not match")
	}

	return u, nil
}

func (s *Server) signInWithOAuthInfo(ctx context.Context, info *OAuthInfo, provider string) (*User, error) {
	u, err := s.userStore.GetByOAuth(ctx, info, provider)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) SignUp(ctx context.Context, in *gonyom.SignUpRequest) (*gonyom.SignReply, error) {
	var u *User
	var err error

	name := in.GetName()
	email := in.GetEmail()
	photourl := in.GetPhotourl()

	switch cred := in.GetCredential().(type) {
	case *gonyom.SignUpRequest_Password:
		u, err = s.signUpWithEmail(ctx, name, email, cred.Password, photourl)
	case *gonyom.SignUpRequest_Oauthinfo:
		provider := in.GetOauthprovider()
		u, err = s.signUpWithOAuthInfo(ctx, name, email, OAuthInfoFromProto(cred.Oauthinfo), provider, photourl)
	default:
		err = errors.NewUnimplementedError("type %v", cred)
	}
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	token, err := s.jwtManager.NewAuthToken(u.Id)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.SignReply{Account: u.ToProto(), Token: token}, nil
}

func (s *Server) SignIn(ctx context.Context, in *gonyom.SignInRequest) (*gonyom.SignReply, error) {
	var u *User
	var err error

	switch cred := in.GetCredential().(type) {
	case *gonyom.SignInRequest_Emailcred:
		u, err = s.signInWithEmail(ctx, cred.Emailcred.Email, cred.Emailcred.Password)
	case *gonyom.SignInRequest_Oauthinfo:
		provider := in.GetOauthprovider()
		u, err = s.signInWithOAuthInfo(ctx, OAuthInfoFromProto(cred.Oauthinfo), provider)
	default:
		err = errors.NewUnimplementedError("type %v", cred)
	}
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	token, err := s.jwtManager.NewAuthToken(u.Id)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.SignReply{Account: u.ToProto(), Token: token}, nil
}

func (s *Server) SignOut(ctx context.Context, in *gonyom.EmptyParams) (*gonyom.EmptyParams, error) {
	return &gonyom.EmptyParams{}, nil
}

func (s *Server) Get(ctx context.Context, request *gonyom.GetAccountRequest) (*gonyom.GetAccountReply, error) {
	uid := ctx.Value(CtxKeyUid).(string)
	if uid == "" {
		return nil, errors.GrpcError(errors.NewAuthenticationError("UID not provided"))
	}
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.GetAccountReply{Account: u.ToProto()}, nil
}

func (s *Server) Update(ctx context.Context, request *gonyom.UpdateAccountRequest) (*gonyom.UpdateAccountReply, error) {
	uid := ctx.Value(CtxKeyUid).(string)
	if uid == "" {
		return nil, errors.GrpcError(errors.NewAuthenticationError("UID not provided"))
	}
	user, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	path := request.GetPath()
	value := request.GetValue()

	if path == "password" {
		value, err = hashPassword(value)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
	}

	err = s.userStore.Update(ctx, user, path, value)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	user, err = s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.UpdateAccountReply{Account: user.ToProto()}, nil
}

func (s *Server) AcceptInvite(ctx context.Context, request *gonyom.AcceptInviteRequest) (*gonyom.AcceptInviteReply, error) {
	uid := ctx.Value(CtxKeyUid).(string)
	pid := request.GetPetId()
	if uid == "" || pid == "" {
		return nil, errors.GrpcError(errors.NewAuthenticationError("UID or PID not provided"))
	}
	_, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	err = s.userStore.AddPet(ctx, uid, pid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	user, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.AcceptInviteReply{Account: user.ToProto()}, nil
}

func (s *Server) Delete(ctx context.Context, request *gonyom.DeleteAccountRequest) (*gonyom.DeleteAccountReply, error) {
	// TODO implement me
	panic("implement me")
}
