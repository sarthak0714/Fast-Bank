package domain

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	Id int `json:"id"`
	jwt.RegisteredClaims
}
