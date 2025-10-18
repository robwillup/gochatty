package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_key") // TODO: Use a secure value.

func GenerateJWT(username string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
