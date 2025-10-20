package auth

import (
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTAndCheckJWT(t *testing.T) {
	tokenString, err := GenerateJWT("bob")
	if err != nil {
		t.Fatalf("Error generating JWT %v", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret_key"), nil
	})

	if err != nil {
		t.Fatalf("Error parsing token %v", err)
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		t.Fatalf("Error validating token: %v", err)
	}

	if claims.Username != "bob" {
		t.Fatalf("Invalid username in claims: %s", claims.Username)
	}
}
