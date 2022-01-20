package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/aiceru/protonyom/gonyom"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"ohmnyom/domain/user"
	"ohmnyom/internal/firestore"
	userdb "ohmnyom/internal/firestore/user"
)

func main() {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8989")
	ctx := context.Background()

	// firestoreClient, err := firestore.NewClient(ctx, "ohmnyom", filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	firestoreClient := firestore.NewEmulatorClient(ctx)
	userService := userdb.NewService(ctx, firestoreClient)

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
