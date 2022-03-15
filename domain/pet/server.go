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
	contentType := request.GetProfileContentType()
	profileImageBytes := request.GetProfilePhoto()

	if profileImageBytes != nil {
		link, err := s.upload(ctx, newPet.NewProfilePath(), profileImageBytes, contentType)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		newPet.Photourl = link
	}

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
	contentType := request.GetProfileContentType()
	profileImageBytes := request.GetProfilePhoto()

	if profileImageBytes != nil {
		link, err := s.upload(ctx, newPet.NewProfilePath(), profileImageBytes, contentType)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		newPet.Photourl = link
	}

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
		Pets: pets.ToProto(),
	}, nil
}

func (s *Server) GetPetWithFeeds(ctx context.Context, request *gonyom.GetPetWithFeedsRequest) (*gonyom.GetPetWithFeedsReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	petId := request.GetPetId()

	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	if !u.HasPet(petId) {
		return nil, errors.GrpcError(errors.NewNotFoundError("pet %v is not belonging to you", petId))
	}

	pet, err := s.petStore.Get(ctx, petId)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	// TODO implement feed store
	// feeds, err := s.feedStore.GetFeeds(ctx, petId)

	return &gonyom.GetPetWithFeedsReply{
		PetFeeds: &gonyom.PetFeeds{
			Pet: pet.ToProto(),
			// TODO: implement feed
			Feeds: nil,
		},
	}, nil
}

func (s *Server) upload(ctx context.Context, path string, content []byte, contentType string) (string, error) {
	mediaLink, err := s.storage.Upload(ctx, &storage.Object{
		RootDir:     googleStorage.RootBucket,
		Path:        path,
		ContentType: contentType,
		Bytes:       content,
	})
	if err != nil {
		return "", errors.NewInternalError("%v", err)
	}
	return mediaLink, nil
}
