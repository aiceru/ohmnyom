package pet

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/aiceru/protonyom/gonyom"
	"github.com/rs/xid"
)

const (
	NameField     = "name"
	PhotourlField = "photourl"
	AdoptedField  = "adopted"
	FamilyField   = "family"
	SpeciesField  = "species"

	storageSep         = "/"
	storageDirPet      = "pets"
	storageDirProfiles = "profiles"
	StorageRoot        = "ohmnyom"
)

type List []*Pet

func (list List) ToProto() []*gonyom.Pet {
	ret := make([]*gonyom.Pet, len(list))
	for i, p := range list {
		ret[i] = p.ToProto()
	}
	return ret
}

type Pet struct {
	Id       string    `firestore:"id"`
	Name     string    `firestore:"name,omitempty"`
	Photourl string    `firestore:"photourl,omitempty"`
	Adopted  time.Time `firestore:"adopted"`
	Family   string    `firestore:"family,omitempty"`
	Species  string    `firestore:"species,omitempty"`
	Feeders  []string  `firestore:"feeders,omitempty"`
}

func IsUpdatableField(field string) bool {
	return field != "Id"
}

func NewPetId() string {
	return xid.New().String()
}

func FromProto(p *gonyom.Pet) *Pet {
	return &Pet{
		Id:       p.Id,
		Name:     p.Name,
		Photourl: p.Photourl,
		Adopted:  time.Unix(p.Adopted, 0),
		Family:   p.Family,
		Species:  p.Species,
		Feeders:  p.Feeders,
	}
}

func (p *Pet) ToProto() *gonyom.Pet {
	return &gonyom.Pet{
		Id:       p.Id,
		Name:     p.Name,
		Photourl: p.Photourl,
		Adopted:  p.Adopted.Unix(),
		Family:   p.Family,
		Species:  p.Species,
		Feeders:  p.Feeders,
	}
}

func (p *Pet) ProfileDir() string {
	return strings.Join([]string{storageDirPet, p.Id, storageDirProfiles}, storageSep)
}

func (p *Pet) NewProfilePath() string {
	timeStr := strconv.FormatInt(time.Now().UTC().UnixNano(), 16)
	return strings.Join([]string{storageDirPet, p.Id, storageDirProfiles, timeStr}, storageSep)
}

type Store interface {
	Get(ctx context.Context, id string) (*Pet, error)
	GetList(ctx context.Context, ids []string) (List, error)
	Put(ctx context.Context, pet *Pet) error
	Update(ctx context.Context, id string, pathValues map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	AddFeeder(ctx context.Context, id, uid string) error
	DeleteFeeder(ctx context.Context, id, uid string) error
}
