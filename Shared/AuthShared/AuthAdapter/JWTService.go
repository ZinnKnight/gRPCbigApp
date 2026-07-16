package AuthAdapter

import (
	"gRPCbigapp/Shared/AuthShared/AuthClaims"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// можно вот эту фигню запихать в Client Service без последствий и смерджить с Policy если потребуется
// Но всё остальное тогда останется

type JWTService struct {
	secretKey []byte
	ttl       time.Duration
}

func NewJWTService(secretKey []byte, ttl time.Duration) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		ttl:       ttl,
	}
}

func (serv *JWTService) TTLinSeconds() int64 {
	return int64(serv.ttl.Seconds())
}

func (serv *JWTService) GenerateToken(userID, userName, userPlan string) (string, error) {
	timeStomp := time.Now()
	claims := AuthClaims.Claims{
		UserID:   userID,
		UserName: userName,
		UserPlan: userPlan,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(timeStomp.Add(serv.ttl)),
			IssuedAt:  jwt.NewNumericDate(timeStomp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(serv.secretKey)
}
