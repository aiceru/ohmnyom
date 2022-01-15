package user

import (
	"context"
	"log"

	"github.com/aiceru/protonyom/gonyom"
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

func (s *Server) SignIn(ctx context.Context, in *gonyom.SignInRequest) (*gonyom.SignInReply, error) {
	log.Printf("Received: %v", in)
	return &gonyom.SignInReply{
		Error: gonyom.ErrorCode_OK,
		Account: &gonyom.Account{
			Id:        "grpc-test-id",
			Email:     "grpc@test.com",
			Name:      "grpcgrpc",
			Photourl:  "https://grpc.com",
			Oauthtype: gonyom.OAuthType_GOOGLE,
			Oauthid:   "grpc-oauth-id",
			Signedup: &gonyom.Timestamp{
				Seconds: 24,
				Nanos:   12,
			},
			Pets: []string{"pet1", "pet2"},
		},
	}, nil
}
