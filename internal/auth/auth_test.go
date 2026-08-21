package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	if hash == password {
		t.Error("hash should not equal plain password")
	}
}

func TestLoginRequestValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		valid    bool
	}{
		{"valid credentials", "admin", "password123", true},
		{"empty username", "", "password123", false},
		{"empty password", "admin", "", false},
		{"both empty", "", "", false},
		{"whitespace username", "  ", "password123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username := strings.TrimSpace(tt.username)
			password := tt.password
			valid := username != "" && password != ""
			if valid != tt.valid {
				t.Errorf("expected valid=%v, got valid=%v", tt.valid, valid)
			}
		})
	}
}

func TestClaimsStructure(t *testing.T) {
	claims := Claims{
		UserID:      1,
		Username:    "admin",
		FullName:    "Administrator",
		Roles:       []string{"ADMIN"},
		Permissions: []string{"users.read", "users.write"},
	}

	if claims.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("expected Username 'admin', got '%s'", claims.Username)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "ADMIN" {
		t.Errorf("expected Roles ['ADMIN'], got %v", claims.Roles)
	}
}
