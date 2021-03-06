package servers

import (
	"context"
	"time"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/domain/feed"
	"ohmnyom/domain/user"
	"ohmnyom/internal/errors"
)

type FeedServer struct {
	feedStore feed.Store
	userStore user.Store
	gonyom.UnimplementedFeedApiServer
}

func NewFeedServer(store feed.Store, userStore user.Store) *FeedServer {
	return &FeedServer{
		feedStore: store,
		userStore: userStore,
	}
}

func (s *FeedServer) AddFeed(ctx context.Context, request *gonyom.AddFeedRequest) (*gonyom.AddFeedReply, error) {
	// uid := ctx.Value(user.CtxKeyUid).(string)
	// TODO user check??
	newFeed, err := feed.NewFromProto(request.GetFeed())
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	feeder, err := s.userStore.Get(ctx, newFeed.FeederId)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	if err := s.feedStore.Put(ctx, newFeed); err != nil {
		return nil, errors.GrpcError(err)
	}

	check, err := s.feedStore.Get(ctx, newFeed.PetId, newFeed.Id)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.AddFeedReply{
		Feed: check.ToProto(feeder.Name),
	}, nil
}

func (s *FeedServer) GetFeeds(ctx context.Context, request *gonyom.GetFeedsRequest) (*gonyom.GetFeedsReply, error) {
	petId := request.GetPetId()
	startAfter := time.Unix(request.GetStartAfter(), 0)
	limit := request.GetLimit()

	feeds, err := s.feedStore.GetFeedsOfPet(ctx, petId, startAfter, int(limit))
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	ret := make([]*gonyom.Feed, len(feeds))
	for i, f := range feeds {
		feeder, err := s.userStore.Get(ctx, f.FeederId)
		if err != nil {
			return nil, err
		}
		ret[i] = f.ToProto(feeder.Name)
	}
	return &gonyom.GetFeedsReply{
		Feeds: ret,
	}, nil
}

func (s *FeedServer) DeleteFeed(ctx context.Context, request *gonyom.DeleteFeedRequest) (*gonyom.DeleteFeedReply, error) {
	petId := request.GetPetId()
	feedId := request.GetFeedId()

	if err := s.feedStore.Delete(ctx, petId, feedId); err != nil {
		return nil, err
	}
	return &gonyom.DeleteFeedReply{}, nil
}

func (s *FeedServer) UpdateFeed(ctx context.Context, request *gonyom.UpdateFeedRequest) (*gonyom.UpdateFeedReply, error) {
	newFeed, err := feed.FromProto(request.GetFeed())
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	feeder, err := s.userStore.Get(ctx, newFeed.FeederId)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	updates := map[string]interface{}{
		"timestamp": newFeed.Timestamp,
		"amount":    newFeed.Amount,
		"unit":      newFeed.Unit,
	}
	if err := s.feedStore.Update(ctx, newFeed.PetId, newFeed.Id, updates); err != nil {
		return nil, errors.GrpcError(err)
	}

	check, err := s.feedStore.Get(ctx, newFeed.PetId, newFeed.Id)
	if err != nil {
		return nil, errors.GrpcError(err)
	}

	return &gonyom.UpdateFeedReply{
		Feed: check.ToProto(feeder.Name),
	}, nil
}
