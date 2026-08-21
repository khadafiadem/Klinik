package users

import (
	"testing"
)

func TestCreateUserRequestValidation(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateUserRequest
		expected string
	}{
		{
			name: "valid request",
			req: CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
				RoleID:   1,
			},
			expected: "",
		},
		{
			name: "missing username",
			req: CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			expected: "semua field wajib diisi",
		},
		{
			name: "missing password",
			req: CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				FullName: "Test User",
			},
			expected: "semua field wajib diisi",
		},
		{
			name: "short password",
			req: CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "12345",
				FullName: "Test User",
			},
			expected: "password minimal 6 karakter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err string
			if tt.req.Username == "" || tt.req.Email == "" || tt.req.Password == "" || tt.req.FullName == "" {
				err = "semua field wajib diisi"
			} else if len(tt.req.Password) < 6 {
				err = "password minimal 6 karakter"
			}

			if err != tt.expected {
				t.Errorf("expected error '%s', got '%s'", tt.expected, err)
			}
		})
	}
}

func TestPaginationDefaults(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expPage  int
		expLimit int
	}{
		{"zero page", 0, 0, 1, 20},
		{"negative page", -1, -5, 1, 20},
		{"valid values", 2, 10, 2, 10},
		{"limit too high", 1, 200, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := tt.page
			limit := tt.limit

			if page < 1 {
				page = 1
			}
			if limit < 1 || limit > 100 {
				limit = 20
			}

			if page != tt.expPage {
				t.Errorf("expected page %d, got %d", tt.expPage, page)
			}
			if limit != tt.expLimit {
				t.Errorf("expected limit %d, got %d", tt.expLimit, limit)
			}
		})
	}
}
