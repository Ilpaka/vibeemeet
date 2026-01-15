package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"video_conference/internal/domain"
	"video_conference/internal/repository"
	"video_conference/pkg/logger"
)

// ExternalAuthMiddleware валидирует JWT токены от внешнего Auth-сервиса (NextUp)
type ExternalAuthMiddleware struct {
	jwtSecret []byte
	userRepo  repository.UserRepository
	log       logger.Logger
}

// ExternalJWTClaims - структура claims от NextUp Auth-сервиса
// NextUp генерирует токены с user_id, email и display_name
type ExternalJWTClaims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	jwt.RegisteredClaims
}

// NewExternalAuthMiddleware создает middleware для внешнего Auth-сервиса
func NewExternalAuthMiddleware(jwtSecret string, userRepo repository.UserRepository, log logger.Logger) *ExternalAuthMiddleware {
	return &ExternalAuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		userRepo:  userRepo,
		log:       log,
	}
}

// RequireAuth требует валидный JWT токен от Auth-сервиса
func (m *ExternalAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.log.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.log.Warn("Invalid Authorization header format", "header", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		m.log.Debug("Parsing JWT token", "token_prefix", tokenString[:20]+"...")
		
		claims, err := m.parseToken(tokenString)
		if err != nil {
			m.log.Warn("Token validation failed", "error", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		
		m.log.Debug("Token validated successfully", "user_id", claims.UserID)

		// Парсим user_id как UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			m.log.Debug("Invalid user_id in token", "user_id", claims.UserID, "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Auto-provisioning: создаем пользователя если его нет
		if err := m.ensureUserExists(c.Request.Context(), userID, claims.Email, claims.DisplayName); err != nil {
			m.log.Error("Failed to ensure user exists", "user_id", userID.String(), "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to provision user"})
			c.Abort()
			return
		}

		// Устанавливаем user_id в контекст (как uuid.UUID для совместимости с handlers)
		c.Set("user_id", userID)
		c.Set("user_id_string", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_display_name", claims.DisplayName)
		
		c.Next()
	}
}

// OptionalAuth проверяет токен если он есть, но не требует его
func (m *ExternalAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := m.parseToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", userID)
		c.Set("user_id_string", claims.UserID)
		c.Next()
	}
}

// parseToken парсит и валидирует JWT токен
func (m *ExternalAuthMiddleware) parseToken(tokenString string) (*ExternalJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ExternalJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ExternalJWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ensureUserExists проверяет существование пользователя в БД и создает если нужно (auto-provisioning)
func (m *ExternalAuthMiddleware) ensureUserExists(ctx context.Context, userID uuid.UUID, email string, displayName string) error {
	// Проверяем существует ли пользователь по ID
	existingUser, err := m.userRepo.GetByID(ctx, userID)
	if err == nil && existingUser != nil {
		// Пользователь существует с этим ID
		m.log.Debug("User already exists in DB", "user_id", userID.String())
		return nil
	}

	// Пользователя нет - создаем (auto-provisioning)
	m.log.Info("Auto-provisioning user from Auth service", "user_id", userID.String(), "email", email)

	now := time.Now()
	newUser := &domain.User{
		ID:              userID,
		Email:           email,
		PasswordHash:    "", // Пароль хранится в Auth-сервисе
		DisplayName:     displayName,
		GlobalRole:      domain.GlobalRoleUser,
		IsActive:        true,
		IsEmailVerified: true, // Предполагаем что Auth-сервис уже проверил
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := m.userRepo.Create(ctx, newUser); err != nil {
		// Проверяем дублирование (race condition с другим запросом)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			// Другой запрос уже создал пользователя - проверим что ID совпадает
			checkUser, checkErr := m.userRepo.GetByID(ctx, userID)
			if checkErr == nil && checkUser != nil {
				m.log.Debug("User was created by concurrent request", "user_id", userID.String())
				return nil
			}
			// Если пользователь с таким email уже существует но с другим ID - это проблема миграции
			// Возвращаем ошибку чтобы администратор мог разобраться
			m.log.Error("User exists with different ID", "user_id", userID.String(), "email", email, "error", err)
			return fmt.Errorf("user with email %s already exists with different ID, please contact administrator", email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	m.log.Info("User auto-provisioned successfully", "user_id", userID.String())
	return nil
}
