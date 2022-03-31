package servers

import (
	"context"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/domain/pet"
	"ohmnyom/domain/user"
	"ohmnyom/i18n"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/storage"
	"ohmnyom/internal/util"
)

type PetServer struct {
	petStore  pet.Store
	userStore user.Store
	storage   storage.Storage
	gonyom.UnimplementedPetApiServer
}

func NewPetServer(store pet.Store, userStore user.Store, storage storage.Storage) *PetServer {
	return &PetServer{
		petStore:  store,
		userStore: userStore,
		storage:   storage,
	}
}

func (s *PetServer) GetFamilies(ctx context.Context, request *gonyom.GetFamiliesRequest) (*gonyom.GetFamiliesReply, error) {
	lang := i18n.SupportOrFallback(request.GetLanguage())
	families := i18n.SupportedFamilies[lang]
	return &gonyom.GetFamiliesReply{
		Families: families,
	}, nil
}

func (s *PetServer) AddPet(ctx context.Context, request *gonyom.AddPetRequest) (*gonyom.AddPetReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	newPet := pet.FromProto(request.GetPet())
	newPet.Id = pet.NewPetId()
	newPet.Feeders = []string{u.Id}

	contentType := request.GetProfileContentType()
	profileImageBytes := request.GetProfilePhoto()

	if profileImageBytes != nil {
		link, err := s.uploadStorage(ctx, newPet.NewProfilePath(), profileImageBytes, contentType)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		newPet.Photourl = link
	}

	if err := s.petStore.Put(ctx, newPet); err != nil {
		return nil, errors.GrpcError(err)
	}
	if err := s.userStore.AddPet(ctx, u.Id, newPet.Id); err != nil {
		return nil, errors.GrpcError(err)
	}

	account, err := s.userStore.Get(ctx, u.Id)
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

func (s *PetServer) UpdatePet(ctx context.Context, request *gonyom.UpdatePetRequest) (*gonyom.UpdatePetReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	newPet := pet.FromProto(request.GetPet())
	contentType := request.GetProfileContentType()
	profileImageBytes := request.GetProfilePhoto()

	if profileImageBytes != nil {
		link, err := s.uploadStorage(ctx, newPet.NewProfilePath(), profileImageBytes, contentType)
		if err != nil {
			return nil, errors.GrpcError(err)
		}
		newPet.Photourl = link
	}

	if err := s.petStore.Update(ctx, newPet.Id, map[string]interface{}{
		pet.NameField:     newPet.Name,
		pet.PhotourlField: newPet.Photourl,
		pet.AdoptedField:  newPet.Adopted,
		pet.FamilyField:   newPet.Family,
		pet.SpeciesField:  newPet.Species,
	}); err != nil {
		return nil, errors.GrpcError(err)
	}

	account, err := s.userStore.Get(ctx, u.Id)
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

func (s *PetServer) DeletePet(ctx context.Context, request *gonyom.DeletePetRequest) (*gonyom.DeletePetReply, error) {
	uid := ctx.Value(user.CtxKeyUid).(string)
	u, err := s.userStore.Get(ctx, uid)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	petId := request.GetPetId()
	if !u.HasPet(petId) {
		return nil, errors.GrpcError(errors.NewNotFoundError("pet id %s from pet list of user %s"))
	}

	p, err := s.petStore.Get(ctx, petId)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	// delete pet id from user doc
	if err := s.userStore.DeletePet(ctx, uid, petId); err != nil {
		return nil, errors.GrpcError(err)
	}

	p.Feeders = util.Remove(p.Feeders, u.Id)
	if len(p.Feeders) == 0 {
		// delete from firestore
		if err := s.petStore.Delete(ctx, petId); err != nil {
			return nil, errors.GrpcError(err)
		}
		// delete profile
		if err := s.storage.DeleteDir(ctx, pet.StorageRoot, p.ProfileDir()); err != nil {
			return nil, errors.GrpcError(err)
		}
	} else {
		// update pet doc
		if err := s.petStore.DeleteFeeder(ctx, petId, uid); err != nil {
			return nil, errors.GrpcError(err)
		}
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

func (s *PetServer) GetPetList(ctx context.Context, request *gonyom.GetPetListRequest) (*gonyom.GetPetListReply, error) {
	petIds := request.GetPetIds()

	pets, err := s.petStore.GetList(ctx, petIds)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.GetPetListReply{
		Pets: pets.ToProto(),
	}, nil
}

func (s *PetServer) GetPet(ctx context.Context, request *gonyom.GetPetRequest) (*gonyom.GetPetReply, error) {
	petId := request.GetPetId()
	p, err := s.petStore.Get(ctx, petId)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.GetPetReply{
		Pet: p.ToProto(),
	}, nil
}

func (s *PetServer) uploadStorage(ctx context.Context, path string, content []byte, contentType string) (string, error) {
	mediaLink, err := s.storage.Upload(ctx, &storage.Object{
		Root:        pet.StorageRoot,
		Path:        path,
		ContentType: contentType,
		Bytes:       content,
	})
	if err != nil {
		return "", errors.NewInternalError("%v", err)
	}
	return mediaLink, nil
}
