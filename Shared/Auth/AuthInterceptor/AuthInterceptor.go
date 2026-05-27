package AuthInterceptor

import (
	"context"
	"gRPCbigapp/Shared/Auth/AuthCTX"
	"gRPCbigapp/Shared/Auth/AuthClaims"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	registration = "/auth.AuthService/UserRegistration"
	login        = "/auth.AuthService/UserLogin"
	pref         = "bearer "
)

var publicMethods = map[string]bool{
	registration: true,
	login:        true,
}

func AuthInterceptor(jwtSecretKey []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
		}
		rawToken := authHeaders[0]
		if len(rawToken) < len(pref) || strings.EqualFold(rawToken[:len(pref)], pref) {
			return nil, status.Errorf(codes.Unauthenticated, "authorization supposed use a Bearer schema")
		}
		tokenStr := strings.TrimSpace(rawToken[len(pref):])

		claims := &AuthClaims.Claims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
			}
			return jwtSecretKey, nil
		},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid auth token")
		}

		if claims.UserID == "" || claims.UserName == "" || claims.UserPlan == "" {
			return nil, status.Error(codes.Unauthenticated, "incomplete claims invalid auth token")
		}

		ctx = AuthCTX.PutUser(ctx, &AuthCTX.UserAuth{
			UserID:   claims.UserID,
			UserName: claims.UserName,
			UserPlan: claims.UserPlan,
		})
		return handler(ctx, req)
	}
}
