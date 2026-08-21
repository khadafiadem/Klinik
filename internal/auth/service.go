package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"klinik-app/internal/logger"
)

type Service struct {
	repo      *Repository
	jwtSecret []byte
	jwtExpiry time.Duration
}

func NewService(db *sql.DB, jwtSecret string, jwtExpiryHours int) *Service {
	return &Service{
		repo:      NewRepository(db),
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: time.Duration(jwtExpiryHours) * time.Hour,
	}
}

type Claims struct {
	UserID      int      `json:"user_id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("username atau password salah")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("akun tidak aktif")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("username atau password salah")
	}

	roles, err := s.repo.GetUserRoles(user.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil role: %w", err)
	}

	permissions, err := s.repo.GetUserPermissions(user.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil permission: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	permNames := make([]string, len(permissions))
	for i, perm := range permissions {
		permNames[i] = perm.Name
	}

	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Roles:       roleNames,
		Permissions: permNames,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "klinik-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token: %w", err)
	}

	if err := s.repo.UpdateLastLogin(user.ID); err != nil {
		logger.Error.Printf("Gagal update last login: %v", err)
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
		User:      *user,
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing tidak valid")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token tidak valid: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token tidak valid")
	}

	return claims, nil
}

func (s *Service) GetUserByID(id int) (*User, error) {
	return s.repo.GetUserByID(id)
}

func (s *Service) GetUserRoles(userID int) ([]Role, error) {
	return s.repo.GetUserRoles(userID)
}

func (s *Service) GetUserPermissions(userID int) ([]Permission, error) {
	return s.repo.GetUserPermissions(userID)
}

func (s *Service) UserExists(username, email string) (bool, error) {
	return s.repo.UserExists(username, email)
}

func (s *Service) AssignRole(userID, roleID int) error {
	return s.repo.AssignRole(userID, roleID)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("gagal hash password: %w", err)
	}
	return string(bytes), nil
}
