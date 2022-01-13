package main

import (
	"context"
	"log"
	"net"

	"github.com/aiceru/protonyom/gonyom"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
)

type Outer struct {
	Number int      `firestore:"number,omitempty"`
	Str    string   `firestore:"str,omitempty"`
	Inners []*Inner `firestore:"inners,omitempty"`
}

type Inner struct {
	Str string `firestore:"test,omitempty"`
}

type server struct {
	gonyom.UnimplementedSignApiServer
}

func (s *server) SignIn(ctx context.Context, in *gonyom.SignInRequest) (*gonyom.SignInReply, error) {
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

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer(grpc.ForceServerCodec(encoding.GetCodec(gzip.Name)))
	gonyom.RegisterSignApiServer(s, &server{})
	log.Println("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
