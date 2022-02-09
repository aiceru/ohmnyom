package pet

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/domain/user"
	"ohmnyom/i18n"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/jwt"
)

type Server struct {
	petStore   Store
	userStore  user.Store
	jwtManager *jwt.Manager
	gonyom.UnimplementedPetApiServer
}

func NewServer(store Store, userStore user.Store, jwtManager *jwt.Manager) *Server {
	return &Server{
		petStore:   store,
		userStore:  userStore,
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

func (s *Server) AddPet(ctx context.Context, request *gonyom.AddPetRequest) (*gonyom.PetListReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	newPet := fromProto(request.GetPet())
	newPet.Id = newPetId()

	if err := s.petStore.Put(ctx, newPet); err != nil {
		return nil, errors.GrpcError(err)
	}
	if err := s.userStore.AddPet(ctx, uid, newPet.Id); err != nil {
		return nil, errors.GrpcError(err)
	}

	pets, err := s.getPetListOf(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.PetListReply{
		Pets: pets.ToProto(),
	}, nil
}

func (s *Server) UpdatePet(ctx context.Context, request *gonyom.UpdatePetRequest) (*gonyom.PetListReply, error) {
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

	pets, err := s.getPetListOf(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.PetListReply{
		Pets: pets.ToProto(),
	}, nil
}

func (s *Server) DeletePet(ctx context.Context, request *gonyom.DeletePetRequest) (*gonyom.PetListReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	oldPetId := request.GetPetId()

	if err := s.petStore.Delete(ctx, oldPetId); err != nil {
		return nil, errors.GrpcError(err)
	}
	if err := s.userStore.DeletePet(ctx, uid, oldPetId); err != nil {
		return nil, errors.GrpcError(err)
	}

	pets, err := s.getPetListOf(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}
	return &gonyom.PetListReply{
		Pets: pets.ToProto(),
	}, nil
}

func (s *Server) GetPetList(ctx context.Context, request *gonyom.GetPetListRequest) (*gonyom.PetListReply, error) {
	petIds := request.GetPetIds()

	pets, err := s.petStore.GetList(ctx, petIds)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.PetListReply{
		Pets: List(pets).ToProto(),
	}, nil
}

func (s *Server) getPetListOf(ctx context.Context, uid string) (List, error) {
	authUser, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	pets, err := s.petStore.GetList(ctx, authUser.Pets)
	if err != nil {
		return nil, err
	}
	return pets, nil
}
