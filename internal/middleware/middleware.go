package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"klinik-app/internal/auth"
	"klinik-app/internal/logger"
)

type Middleware struct {
	authService *auth.Service
}

func New(authService *auth.Service) *Middleware {
	return &Middleware{authService: authService}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, http.StatusUnauthorized, "Token autentikasi diperlukan")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			sendError(w, http.StatusUnauthorized, "Format token tidak valid")
			return
		}

		tokenString := parts[1]
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			sendError(w, http.StatusUnauthorized, "Token tidak valid atau sudah expired")
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*auth.Claims)
			if !ok {
				sendError(w, http.StatusUnauthorized, "Tidak terautentikasi")
				return
			}

			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range claims.Roles {
					if strings.EqualFold(requiredRole, userRole) {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				sendError(w, http.StatusForbidden, "Tidak memiliki akses ke resource ini")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*auth.Claims)
			if !ok {
				sendError(w, http.StatusUnauthorized, "Tidak terautentikasi")
				return
			}

			hasPermission := false
			for _, requiredPerm := range permissions {
				for _, userPerm := range claims.Permissions {
					if requiredPerm == userPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				sendError(w, http.StatusForbidden, "Tidak memiliki izin yang cukup")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		claims, _ := r.Context().Value("claims").(*auth.Claims)
		userID := "anonymous"
		if claims != nil {
			userID = claims.Username
		}

		logger.Info.Printf("%s %s %d %s %s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start).Round(time.Millisecond), userID)
	})
}

type loginAttempt struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	maxTry   int
	window   time.Duration
}

func NewRateLimiter(maxTry int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string]*loginAttempt),
		maxTry:   maxTry,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	if rl.maxTry <= 0 {
		return false
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	a, exists := rl.attempts[key]
	if !exists {
		rl.attempts[key] = &loginAttempt{count: 1, lastSeen: time.Now()}
		return true
	}

	if time.Since(a.lastSeen) > rl.window {
		a.count = 1
		a.lastSeen = time.Now()
		return true
	}

	if a.count >= rl.maxTry {
		return false
	}

	a.count++
	a.lastSeen = time.Now()
	return true
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
