package AuthAdapter

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
	claims := jwt.MapClaims{
		"uuid":      userID,
		"user_name": userName,
		"user_plan": userPlan,
		"exp":       time.Now().Add(serv.ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(serv.secretKey)
}
