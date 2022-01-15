package main

import (
	"context"
	"log"
	"net"

	"github.com/aiceru/protonyom/gonyom"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"ohmnyom/domain/user"
	userStore "ohmnyom/internal/firestore/user"
)

func main() {
	ctx := context.Background()
	userService := userStore.NewService(ctx, "ohmnyom", "ohmnyom-77df675cb827.json")
	userServer := user.NewServer(userService)

	s := grpc.NewServer(grpc.ForceServerCodec(encoding.GetCodec(gzip.Name)))
	gonyom.RegisterSignApiServer(s, userServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
