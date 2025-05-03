package models

import "github.com/golang-jwt/jwt/v5"

type AppClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}
