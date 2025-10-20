package auth

import "testing"

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "password"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if !CheckPasswordHash(password, hashed) {
		t.Fatalf("Hashed password does not match")
	}
}
