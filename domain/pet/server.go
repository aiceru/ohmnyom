package pet

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/domain/user"
	"ohmnyom/i18n"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/jwt"
	"ohmnyom/internal/storage"
	"ohmnyom/internal/storage/googleStorage"
)

type Server struct {
	petStore   Store
	userStore  user.Store
	storage    storage.Storage
	jwtManager *jwt.Manager
	gonyom.UnimplementedPetApiServer
}

func NewServer(store Store, userStore user.Store, storage storage.Storage, jwtManager *jwt.Manager) *Server {
	return &Server{
		petStore:   store,
		userStore:  userStore,
		storage:    storage,
		jwtManager: jwtManager,
	}
}

func (s *Server) GetFamilies(ctx context.Context, request *gonyom.GetFamiliesRequest) (*gonyom.GetFamiliesReply, error) {
	lang := i18n.SupportOrFallback(request.GetLanguage())
	families := i18n.SupportedFamilies[lang]
	return &gonyom.GetFamiliesReply{
		Families: families,
	}, nil
}

func (s *Server) AddPet(ctx context.Context, request *gonyom.AddPetRequest) (*gonyom.AddPetReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	newPet := fromProto(request.GetPet())
	newPet.Id = newPetId()

	if err := s.petStore.Put(ctx, newPet); err != nil {
		return nil, errors.GrpcError(err)
	}
	if err := s.userStore.AddPet(ctx, uid, newPet.Id); err != nil {
		return nil, errors.GrpcError(err)
	}

	account, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	pets, err := s.petStore.GetList(ctx, account.Pets)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.AddPetReply{
		Account: account.ToProto(),
		Pets:    pets.ToProto(),
	}, nil
}

func (s *Server) UpdatePet(ctx context.Context, request *gonyom.UpdatePetRequest) (*gonyom.UpdatePetReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	newPet := fromProto(request.GetPet())

	if err := s.petStore.Update(ctx, newPet.Id, map[string]interface{}{
		nameField:     newPet.Name,
		photourlField: newPet.Photourl,
		adoptedField:  newPet.Adopted,
		familyField:   newPet.Family,
		speciesField:  newPet.Species,
	}); err != nil {
		return nil, errors.GrpcError(err)
	}

	account, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	pets, err := s.petStore.GetList(ctx, account.Pets)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.UpdatePetReply{
		Pets: pets.ToProto(),
	}, nil
}

func (s *Server) DeletePet(ctx context.Context, request *gonyom.DeletePetRequest) (*gonyom.DeletePetReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	oldPetId := request.GetPetId()

	if err := s.petStore.Delete(ctx, oldPetId); err != nil {
		return nil, errors.GrpcError(err)
	}
	if err := s.userStore.DeletePet(ctx, uid, oldPetId); err != nil {
		return nil, errors.GrpcError(err)
	}

	account, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	pets, err := s.petStore.GetList(ctx, account.Pets)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.DeletePetReply{
		Account: account.ToProto(),
		Pets:    pets.ToProto(),
	}, nil
}

func (s *Server) GetPetList(ctx context.Context, request *gonyom.GetPetListRequest) (*gonyom.GetPetListReply, error) {
	petIds := request.GetPetIds()

	pets, err := s.petStore.GetList(ctx, petIds)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.GetPetListReply{
		Pets: List(pets).ToProto(),
	}, nil
}

func (s *Server) UploadPetProfile(ctx context.Context, request *gonyom.UploadPetProfileRequest) (*gonyom.UploadPetProfileReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	petId := request.GetPetId()
	contentType := request.GetContentType()
	content := request.GetContent()

	user, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	pet, err := s.petStore.Get(ctx, petId)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	mediaLink, err := s.storage.Upload(ctx, &storage.Object{
		RootDir:     googleStorage.RootBucket,
		Path:        pet.NewProfilePath(),
		ContentType: contentType,
		Bytes:       content,
	})
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	if err := s.petStore.Update(ctx, petId, map[string]interface{}{photourlField: mediaLink}); err != nil {
		return nil, errors.GrpcError(err)
	}
	pets, err := s.petStore.GetList(ctx, user.Pets)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.UploadPetProfileReply{
		Pets: pets.ToProto(),
	}, nil
}
