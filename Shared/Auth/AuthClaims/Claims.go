package AuthClaims

import "github.com/golang-jwt/jwt/v5"

// Выделил в отдельный пакет claims что бы оба сервиса могли через него сверять ключи, так поправим проблему проверки что была до этого

type Claims struct {
	UserID   string `json:"uuid"`
	UserName string `json:"user_name"`
	UserPlan string `json:"user_plan"`
	jwt.RegisteredClaims
}
