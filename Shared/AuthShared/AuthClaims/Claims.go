package AuthClaims

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	UserPlan string `json:"user_plan"`
	jwt.RegisteredClaims
}
