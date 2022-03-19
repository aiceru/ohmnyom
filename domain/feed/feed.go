package feed

import (
	"context"
	"time"

	"github.com/aiceru/protonyom/gonyom"
	"github.com/rs/xid"
	"ohmnyom/internal/errors"
)

var updatableFields = []string{"timestamp", "amount", "unit"}

func IsUpdatableField(field string) bool {
	for _, f := range updatableFields {
		if f == field {
			return true
		}
	}
	return false
}

type Feed struct {
	Id        string    `firestore:"id,omitempty"`
	PetId     string    `firestore:"petId,omitempty"`
	Timestamp time.Time `firestore:"timestamp,omitempty"`
	FeederId  string    `firestore:"feederId,omitempty"`
	Amount    float64   `firestore:"amount,omitempty"`
	Unit      string    `firestore:"unit,omitempty"`
}

func newFeedId() string {
	return xid.NewWithTime(time.Now().UTC()).String()
}

// NewFromProto : returns with nil Feeder, have to manually fill it
func NewFromProto(feed *gonyom.Feed) (*Feed, error) {
	if feed.Id != "" {
		return nil, errors.NewInvalidParamError("feed already has ID, %v", feed.Id)
	}
	return &Feed{
		Id:        newFeedId(),
		PetId:     feed.PetId,
		Timestamp: time.Unix(feed.Timestamp, 0),
		FeederId:  feed.FeederId,
		Amount:    feed.Amount,
		Unit:      feed.Unit,
	}, nil
}

// FromProto : returns with nil Feeder, have to manually fill it
func FromProto(feed *gonyom.Feed) (*Feed, error) {
	if feed.Id == "" {
		return nil, errors.NewInvalidParamError("feed does not have ID")
	}
	return &Feed{
		Id:        feed.Id,
		PetId:     feed.PetId,
		Timestamp: time.Unix(feed.Timestamp, 0),
		FeederId:  feed.FeederId,
		Amount:    feed.Amount,
		Unit:      feed.Unit,
	}, nil
}

type List []*Feed

func (f *Feed) ToProto(feederName string) *gonyom.Feed {
	return &gonyom.Feed{
		Id:         f.Id,
		PetId:      f.PetId,
		Timestamp:  f.Timestamp.Unix(),
		FeederId:   f.FeederId,
		FeederName: feederName,
		Amount:     f.Amount,
		Unit:       f.Unit,
	}
}

type Store interface {
	Get(ctx context.Context, petId, feedId string) (*Feed, error)
	GetFeedsOfPet(ctx context.Context, petId string, startAfter time.Time, limit int) ([]*Feed, error)
	Put(ctx context.Context, feed *Feed) error
	Update(ctx context.Context, petId, feedId string, pathValues map[string]interface{}) error
	Delete(ctx context.Context, petId, feedId string) error
}
