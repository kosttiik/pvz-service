package models

import "github.com/golang-jwt/jwt"

type Claims struct {
	UserID string `json:"userId"`
	Role   Role   `json:"role"`
	jwt.StandardClaims
}
