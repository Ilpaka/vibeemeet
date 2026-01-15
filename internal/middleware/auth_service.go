package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthServiceMiddleware проверяет JWT токены от Auth-сервиса
type AuthServiceMiddleware struct {
	jwtSecret []byte
}

// NewAuthServiceMiddleware создает новый middleware для Auth-сервиса
func NewAuthServiceMiddleware(jwtSecret string) *AuthServiceMiddleware {
	return &AuthServiceMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

// JWTClaims представляет структуру claims из Auth-сервиса
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   string   `json:"token_type"`
	IsGuest     bool     `json:"is_guest"`
	jwt.RegisteredClaims
}

// RequireAuth требует аутентификацию (JWT токен от Auth-сервиса)
func (m *AuthServiceMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		
		// Проверяем формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}
		
		tokenString := parts[1]
		
		// Парсим и проверяем токен
		claims, err := m.parseToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}
		
		// Добавляем данные пользователя в context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "display_name", claims.DisplayName)
		ctx = context.WithValue(ctx, "roles", claims.Roles)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)
		ctx = context.WithValue(ctx, "is_guest", claims.IsGuest)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth проверяет токен если он есть, но не требует его
func (m *AuthServiceMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Нет токена - ок, продолжаем без авторизации
			next.ServeHTTP(w, r)
			return
		}
		
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Неверный формат - игнорируем
			next.ServeHTTP(w, r)
			return
		}
		
		tokenString := parts[1]
		claims, err := m.parseToken(tokenString)
		if err != nil {
			// Невалидный токен - игнорируем
			next.ServeHTTP(w, r)
			return
		}
		
		// Добавляем данные в context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "display_name", claims.DisplayName)
		ctx = context.WithValue(ctx, "roles", claims.Roles)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)
		ctx = context.WithValue(ctx, "is_guest", claims.IsGuest)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ForbidGuests запрещает доступ гостям
func (m *AuthServiceMiddleware) ForbidGuests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isGuest, ok := r.Context().Value("is_guest").(bool)
		if ok && isGuest {
			http.Error(w, "Guest access not allowed for this operation", http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// RequirePermission требует определенное право доступа
func (m *AuthServiceMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value("permissions").([]string)
			if !ok {
				http.Error(w, "No permissions found", http.StatusForbidden)
				return
			}
			
			hasPermission := false
			for _, p := range permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}
			
			if !hasPermission {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// parseToken парсит и валидирует JWT токен
func (m *AuthServiceMiddleware) parseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token claims")
}

// Helper functions для извлечения данных из context

// GetUserID извлекает user_id из context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetEmail извлекает email из context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value("email").(string)
	return email, ok
}

// GetDisplayName извлекает display_name из context
func GetDisplayName(ctx context.Context) (string, bool) {
	displayName, ok := ctx.Value("display_name").(string)
	return displayName, ok
}

// IsGuest проверяет, является ли пользователь гостем
func IsGuest(ctx context.Context) bool {
	isGuest, ok := ctx.Value("is_guest").(bool)
	return ok && isGuest
}

// HasPermission проверяет, есть ли у пользователя право
func HasPermission(ctx context.Context, permission string) bool {
	permissions, ok := ctx.Value("permissions").([]string)
	if !ok {
		return false
	}
	
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	
	return false
}
