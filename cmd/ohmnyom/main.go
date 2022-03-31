package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aiceru/protonyom/gonyom"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"ohmnyom/cmd/ohmnyom/servers"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gcpCredentialJsonPath := filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json")
	firestoreClient, err := firestore.NewClient(ctx, "ohmnyom", filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json"))
	if err != nil {
		log.Fatal(err)
	}

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

	userServer := servers.NewUserServer(userStore, petStore, storage, jwtManager)
	petServer := servers.NewPetServer(petStore, userStore, storage)
	feedServer := servers.NewFeedServer(feedStore, userStore)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.ForceServerCodec(encoding.GetCodec(gzip.Name)),
	)
	gonyom.RegisterSignApiServer(grpcServer, userServer)
	gonyom.RegisterAccountApiServer(grpcServer, userServer)
	gonyom.RegisterPetApiServer(grpcServer, petServer)
	gonyom.RegisterFeedApiServer(grpcServer, feedServer)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
