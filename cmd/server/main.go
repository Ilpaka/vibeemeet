package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"video_conference/internal/config"
	"video_conference/internal/handler"
	"video_conference/internal/middleware"
	"video_conference/internal/repository"
	"video_conference/internal/service"
	"video_conference/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	appLogger := logger.New(cfg.Log.Level)

	// Подключение к PostgreSQL
	dbPool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer dbPool.Close()

	// Проверка подключения к БД
	if err := dbPool.Ping(context.Background()); err != nil {
		appLogger.Fatal("Failed to ping database", "error", err)
	}
	appLogger.Info("Database connection established")

	// Подключение к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Проверка подключения к Redis
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		appLogger.Fatal("Failed to connect to Redis", "error", err)
	}
	appLogger.Info("Redis connection established")

	// Инициализация репозиториев
	repos := repository.NewRepositories(dbPool, rdb, appLogger)

	// Инициализация сервисов
	services := service.NewServices(repos, cfg, appLogger)

	// Инициализация middleware
	authMiddleware := middleware.NewAuthMiddleware(services.Auth, appLogger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(services.RateLimit, appLogger)
	participantMiddleware := middleware.ParticipantMiddleware()

	// Инициализация handlers
	handlers := handler.NewHandlers(services, repos, cfg, appLogger)

	// Настройка роутера
	router := setupRouter(handlers, authMiddleware, rateLimitMiddleware, participantMiddleware, cfg, appLogger)

	// Запуск HTTP сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		appLogger.Info("Starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server exited")
}

func setupRouter(
	handlers *handler.Handlers,
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	participantMiddleware gin.HandlerFunc,
	cfg *config.Config,
	log logger.Logger,
) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())

	// Health check
	router.GET("/health", handlers.Health.Check)

	// Server info - для получения IP и настроек сервера
	router.GET("/server-info", handlers.Health.ServerInfo)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Публичные endpoints
		public := v1.Group("/auth")
		{
			public.POST("/register", rateLimitMiddleware.Limit(), handlers.Auth.Register)
			public.POST("/login", rateLimitMiddleware.Limit(), handlers.Auth.Login)
			public.POST("/refresh", handlers.Auth.RefreshToken)
		}

		// Анонимные endpoints (без аутентификации, но с participant_id)
		anonymous := v1.Group("")
		anonymous.Use(participantMiddleware)
		{
			// Анонимные комнаты
			if handlers.AnonymousRoom != nil {
				anonymous.POST("/rooms", handlers.AnonymousRoom.Create)
				anonymous.GET("/rooms/:id", handlers.AnonymousRoom.GetByID)
				anonymous.POST("/rooms/:id/join", handlers.AnonymousRoom.Join)
				anonymous.POST("/rooms/:id/leave", handlers.AnonymousRoom.Leave)
				anonymous.GET("/rooms/:id/participants", handlers.AnonymousRoom.GetParticipants)
			}

			// Анонимный endpoint для получения LiveKit токена
			if handlers.AnonymousMedia != nil {
				anonymous.POST("/rooms/:id/media/token", handlers.AnonymousMedia.GetToken)
			}

			// Анонимный чат
			if handlers.AnonymousChat != nil {
				anonymous.GET("/rooms/:id/chat/messages", handlers.AnonymousChat.GetMessages)
				anonymous.POST("/rooms/:id/chat/messages", handlers.AnonymousChat.SendMessage)
				anonymous.DELETE("/rooms/:id/chat/messages/:messageId", handlers.AnonymousChat.DeleteMessage)
			}
		}

		// Защищенные endpoints
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Пользователи
			users := protected.Group("/users")
			{
				users.GET("/me", handlers.User.GetMe)
				users.PUT("/me", handlers.User.UpdateMe)
				users.GET("/me/settings", handlers.User.GetSettings)
				users.PUT("/me/settings", handlers.User.UpdateSettings)
			}

			// Комнаты (только для аутентифицированных пользователей - устаревшие endpoints)
			// Основные операции теперь через анонимные endpoints
			rooms := protected.Group("/rooms")
			{
				// rooms.POST("", handlers.Room.Create) // Убрано - используем анонимный endpoint
				rooms.GET("", handlers.Room.List)
				// rooms.GET("/:id", handlers.Room.GetByID) // Убрано - используем анонимный endpoint
				rooms.PUT("/:id", handlers.Room.Update)
				rooms.DELETE("/:id", handlers.Room.Delete)
				// rooms.POST("/:id/join", handlers.Room.Join) // Убрано - используем анонимный endpoint
				// rooms.POST("/:id/leave", handlers.Room.Leave) // Убрано - используем анонимный endpoint
				rooms.POST("/:id/invite", handlers.Room.CreateInvite)
				// rooms.GET("/:id/participants", handlers.Room.GetParticipants) // Убрано - используем анонимный endpoint
			}

			// Waiting room
			waitingRoom := protected.Group("/rooms/:id/waiting-room")
			{
				waitingRoom.GET("", handlers.WaitingRoom.List)
				waitingRoom.POST("/:entryId/approve", handlers.WaitingRoom.Approve)
				waitingRoom.POST("/:entryId/reject", handlers.WaitingRoom.Reject)
			}

			// Чат - теперь используем анонимный endpoint
			// chat := protected.Group("/rooms/:id/chat")
			// {
			// 	chat.GET("/messages", handlers.Chat.GetMessages)
			// 	chat.POST("/messages", handlers.Chat.SendMessage)
			// 	chat.PUT("/messages/:messageId", handlers.Chat.EditMessage)
			// 	chat.DELETE("/messages/:messageId", handlers.Chat.DeleteMessage)
			// }

			// Медиа (LiveKit токены) - убрано, используем анонимный endpoint
			// media := protected.Group("/rooms/:id/media")
			// {
			// 	media.POST("/token", handlers.Media.GetToken)
			// }

			// Статистика
			stats := protected.Group("/rooms/:id/stats")
			{
				stats.GET("", handlers.Stats.GetRoomStats)
				stats.GET("/participants/:participantId", handlers.Stats.GetParticipantStats)
			}
		}
	}

	// WebSocket endpoint для чата
	router.GET("/ws/chat/:id", handlers.WebSocket.HandleChat)

	// Screen share endpoints (публичные для демонстрации)
	screenShare := router.Group("/screen-share")
	{
		screenShare.POST("/offer", handlers.ScreenShare.HandleOffer)
		screenShare.POST("/ice/:id", handlers.ScreenShare.HandleICE)
		screenShare.GET("/ice/:id", handlers.ScreenShare.GetICE) // Получение ICE candidates от сервера
		screenShare.POST("/hangup/:id", handlers.ScreenShare.HandleHangup)
		screenShare.GET("/", handlers.ScreenShare.ServeHTML)
	}

	return router
}
