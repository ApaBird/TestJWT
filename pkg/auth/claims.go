package auth

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	jwt.StandardClaims
	Ip string `json:"ip"`
}
