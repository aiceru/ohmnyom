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
	userstore "ohmnyom/internal/firestore/user"
	"ohmnyom/internal/interceptor"
	"ohmnyom/internal/jwt"
)

type CtxKeyType string

func main() {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8989")
	ctx := context.Background()

	// firestoreClient, err := firestore.NewClient(ctx, "ohmnyom", filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	firestoreClient := firestore.NewEmulatorClient(ctx)

	jwtManager := jwt.NewManager([]byte("alsdfkjas;dflkjw;elkfj;ldkfjsdlf"))
	authInterceptor := interceptor.NewAuthInterceptor(jwtManager, "/protonyom.SignApi/SignUp")
	userStore := userstore.New(ctx, firestoreClient)

	userServer := user.NewServer(userStore)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.ForceServerCodec(encoding.GetCodec(gzip.Name)),
	)
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
