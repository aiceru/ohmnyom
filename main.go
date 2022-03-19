package main

import (
	"context"
	"log"
	"net"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aiceru/protonyom/gonyom"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"ohmnyom/domain/feed"
	"ohmnyom/domain/pet"
	"ohmnyom/domain/user"
	"ohmnyom/internal/firestore"
	feedstore "ohmnyom/internal/firestore/feed"
	petstore "ohmnyom/internal/firestore/pet"
	userstore "ohmnyom/internal/firestore/user"
	"ohmnyom/internal/interceptor"
	"ohmnyom/internal/jwt"
	"ohmnyom/internal/path"
	"ohmnyom/internal/storage/googleStorage"
)

func printAddress() {
	nif, _ := net.InterfaceByName("en0")
	addrs, _ := nif.Addrs()
	for _, addr := range addrs {
		re, _ := regexp.Compile(`^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`)
		if re.MatchString(strings.Split(addr.String(), "/")[0]) {
			log.Printf("server address is: %v\n", addr)
		}
	}
}

func main() {
	ctx := context.Background()
	gcpCredentialJsonPath := filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json")

	// firestoreClient, err := firestore.NewClient(ctx, "ohmnyom", filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	firestore.KillEmulator()
	firestore.RunEmulator()
	defer firestore.KillEmulator()
	firestoreClient := firestore.NewEmulatorClient(ctx)

	jwtManager := jwt.NewManager([]byte("temp-test-secret"))
	authInterceptor := interceptor.NewAuthInterceptor(
		jwtManager,
		"/protonyom.SignApi/SignUp",
		"/protonyom.SignApi/SignIn",
		"/protonyom.PetApi/GetFamilies",
	)
	userStore := userstore.New(ctx, firestoreClient)
	petStore := petstore.New(ctx, firestoreClient)
	feedStore := feedstore.New(ctx, firestoreClient)
	storage := googleStorage.New(ctx, gcpCredentialJsonPath)

	userServer := user.NewServer(userStore, jwtManager)
	petServer := pet.NewServer(petStore, userStore, storage)
	feedServer := feed.NewServer(feedStore, userStore)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.ForceServerCodec(encoding.GetCodec(gzip.Name)),
	)
	gonyom.RegisterSignApiServer(grpcServer, userServer)
	gonyom.RegisterAccountApiServer(grpcServer, userServer)
	gonyom.RegisterPetApiServer(grpcServer, petServer)
	gonyom.RegisterFeedApiServer(grpcServer, feedServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
