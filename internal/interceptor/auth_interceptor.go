package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"ohmnyom/domain/user"
	"ohmnyom/internal/errors"
	"ohmnyom/internal/jwt"
)

type AuthInterceptor struct {
	jwtManager *jwt.Manager
	bypass     map[string]struct{}
}

func NewAuthInterceptor(jwtManager *jwt.Manager, bypassMethods ...string) *AuthInterceptor {
	if jwtManager == nil {
		return nil
	}
	bypass := make(map[string]struct{})
	for _, method := range bypassMethods {
		bypass[method] = struct{}{}
	}
	return &AuthInterceptor{
		jwtManager: jwtManager,
		bypass:     bypass,
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.NewAuthenticationError("metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", errors.NewAuthenticationError("authorization token is not provided")
	}

	accessToken := values[0]
	uid, err := i.jwtManager.Verify(accessToken)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, ok := i.bypass[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		uid, err := i.authorize(ctx)
		if err != nil {
			return nil, errors.GrpcError(err)
		}

		c := context.WithValue(ctx, user.CtxKeyUid, uid)
		return handler(c, req)
	}
}
