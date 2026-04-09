package AuthInterceptor

import (
	"context"
	"gRPCbigapp/App/Shared/Auth/AuthCTX"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/auth.AuthService/UserRegistration": true,
	"/auth.AuthService/UserLogin":        true,
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

		tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})
		if err != nil || !token.Valid {
			return nil, status.Errorf(codes.Unauthenticated, "token is not valid")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "token claims not valid")
		}
		user := &AuthCTX.UserAuth{
			UserID:   claimsToString(claims, "user_id"),
			UserName: claimsToString(claims, "user_name"),
			UserPlan: claimsToString(claims, "user_plan"),
		}
		if user.UserID == "" || user.UserName == "" || user.UserPlan == "" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid user id/user name/plan")
		}
		ctx = AuthCTX.PutUser(ctx, user)

		return handler(ctx, req)
	}
}

func claimsToString(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}
