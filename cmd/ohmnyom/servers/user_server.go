package servers

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/domain/pet"
	"ohmnyom/domain/user"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/jwt"
	"ohmnyom/internal/storage"
)

type UserServer struct {
	userStore  user.Store
	petStore   pet.Store
	storage    storage.Storage
	jwtManager *jwt.Manager
	gonyom.UnimplementedSignApiServer
	gonyom.UnimplementedAccountApiServer
}

func NewUserServer(store user.Store, petStore pet.Store, storage storage.Storage, jwtManager *jwt.Manager) *UserServer {
	return &UserServer{
		userStore:  store,
		petStore:   petStore,
		storage:    storage,
		jwtManager: jwtManager,
	}
}

func (s *UserServer) signUpWithEmail(
	ctx context.Context, name, email, password, photourl string) (*user.User, error) {
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

	hashed, err := user.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u, err := user.NewUser(name, email, hashed, nil, photourl)
	if err != nil {
		return nil, err
	}

	if err := s.userStore.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *UserServer) signUpWithOAuthInfo(
	ctx context.Context, name, email string, info *user.OAuthInfo, provider, photourl string) (
	*user.User, error) {
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

	u, err := user.NewUser(name, email, "", map[string]*user.OAuthInfo{provider: info}, photourl)
	if err != nil {
		return nil, err
	}

	if err := s.userStore.Put(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *UserServer) signInWithEmail(ctx context.Context, email, password string) (*user.User, error) {
	u, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if u.Password == "" {
		return nil, errors.NewAuthenticationError("password not set")
	}

	if err := user.CompareHashAndPassword(u.Password, password); err != nil {
		return nil, errors.NewAuthenticationError("password not match")
	}

	return u, nil
}

func (s *UserServer) signInWithOAuthInfo(ctx context.Context, info *user.OAuthInfo, provider string) (*user.User, error) {
	u, err := s.userStore.GetByOAuth(ctx, info, provider)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *UserServer) SignUp(ctx context.Context, in *gonyom.SignUpRequest) (*gonyom.SignReply, error) {
	var u *user.User
	var err error

	name := in.GetName()
	email := in.GetEmail()
	photourl := in.GetPhotourl()

	switch cred := in.GetCredential().(type) {
	case *gonyom.SignUpRequest_Password:
		u, err = s.signUpWithEmail(ctx, name, email, cred.Password, photourl)
	case *gonyom.SignUpRequest_Oauthinfo:
		provider := in.GetOauthprovider()
		u, err = s.signUpWithOAuthInfo(ctx, name, email, user.OAuthInfoFromProto(cred.Oauthinfo), provider, photourl)
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

func (s *UserServer) SignIn(ctx context.Context, in *gonyom.SignInRequest) (*gonyom.SignReply, error) {
	var u *user.User
	var err error

	switch cred := in.GetCredential().(type) {
	case *gonyom.SignInRequest_Emailcred:
		u, err = s.signInWithEmail(ctx, cred.Emailcred.Email, cred.Emailcred.Password)
	case *gonyom.SignInRequest_Oauthinfo:
		provider := in.GetOauthprovider()
		u, err = s.signInWithOAuthInfo(ctx, user.OAuthInfoFromProto(cred.Oauthinfo), provider)
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

func (s *UserServer) SignOut(ctx context.Context, in *gonyom.EmptyParams) (*gonyom.EmptyParams, error) {
	return &gonyom.EmptyParams{}, nil
}

func (s *UserServer) Get(ctx context.Context, request *gonyom.GetAccountRequest) (*gonyom.GetAccountReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	if uid == "" {
		return nil, errors.GrpcError(errors.NewAuthenticationError("UID not provided"))
	}
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.GetAccountReply{Account: u.ToProto()}, nil
}

func (s *UserServer) Update(ctx context.Context, request *gonyom.UpdateAccountRequest) (*gonyom.UpdateAccountReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	if uid == "" {
		return nil, errors.GrpcError(errors.NewAuthenticationError("UID not provided"))
	}
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	path := request.GetPath()
	value := request.GetValue()

	if path == "password" {
		value, err = user.HashPassword(value)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
	}

	err = s.userStore.Update(ctx, u, path, value)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	u, err = s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.UpdateAccountReply{Account: u.ToProto()}, nil
}

func (s *UserServer) AcceptInvite(ctx context.Context, request *gonyom.AcceptInviteRequest) (*gonyom.AcceptInviteReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
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
	err = s.petStore.AddFeeder(ctx, pid, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.AcceptInviteReply{Account: u.ToProto()}, nil
}

func (s *UserServer) UploadProfile(ctx context.Context, request *gonyom.UploadProfileRequest) (*gonyom.UploadProfileResponse, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	contentType := request.GetProfileContentType()
	profileImageBytes := request.GetProfilePhoto()

	if profileImageBytes == nil || len(profileImageBytes) < 1 {
		return nil, errors.GrpcError(errors.NewInvalidParamError("profileImageByte empty"))
	}

	object := &storage.Object{
		Root:        user.StorageRoot,
		Path:        u.NewProfilePath(),
		ContentType: contentType,
		Bytes:       profileImageBytes,
	}
	mediaLink, err := s.storage.Upload(ctx, object)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	if err := s.userStore.Update(ctx, u, "photourl", mediaLink); err != nil {
		return nil, errors.GrpcError(err)
	}

	u, err = s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.UploadProfileResponse{Account: u.ToProto()}, nil
}

func (s *UserServer) Delete(ctx context.Context, request *gonyom.DeleteAccountRequest) (*gonyom.DeleteAccountReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	if uid != request.GetId() {
		return nil, errors.GrpcError(errors.New("cannot delete other user, %s / %s", uid, request.GetId()))
	}
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	pets, err := s.petStore.GetList(ctx, u.Pets)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	for _, p := range pets {
		if err := s.petStore.DeleteFeeder(ctx, p.Id, u.Id); err != nil {
			return nil, errors.GrpcError(err)
		}
	}

	check, err := s.petStore.GetList(ctx, u.Pets)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	for _, p := range check {
		if len(p.Feeders) == 0 {
			if err := s.petStore.Delete(ctx, p.Id); err != nil {
				return nil, errors.GrpcError(err)
			}
		}
	}

	if err := s.userStore.Delete(ctx, u.Id); err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.DeleteAccountReply{}, nil
}
