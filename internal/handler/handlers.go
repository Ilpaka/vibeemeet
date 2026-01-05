package handler

import (
	"video_conference/internal/config"
	"video_conference/internal/repository"
	"video_conference/internal/service"
	"video_conference/pkg/logger"
)

type Handlers struct {
	Health           *HealthHandler
	Auth             *AuthHandler
	User             *UserHandler
	Room             *RoomHandler
	WaitingRoom      *WaitingRoomHandler
	Chat             *ChatHandler
	AnonymousChat    *AnonymousChatHandler
	Media            *MediaHandler
	AnonymousMedia   *AnonymousMediaHandler
	AnonymousRoom    *AnonymousRoomHandler
	Stats            *StatsHandler
	WebSocket        *WebSocketHandler
	ScreenShare      *ScreenShareHandler
}

func NewHandlers(services *service.Services, repos *repository.Repositories, cfg *config.Config, log logger.Logger) *Handlers {
	handlers := &Handlers{
		Health:      NewHealthHandler(cfg),
		Auth:        NewAuthHandler(services.Auth, log),
		User:        NewUserHandler(services.User, log),
		Room:        NewRoomHandler(services.Room, log),
		WaitingRoom: NewWaitingRoomHandler(services.Room, log),
		Chat:        NewChatHandler(services.Chat, log),
		Media:       NewMediaHandler(services.Media, log),
		Stats:       NewStatsHandler(services.Stats, log),
		WebSocket:   NewWebSocketHandler(services.Chat, log),
		ScreenShare: NewScreenShareHandler(services.ScreenCapture, services.AudioCapture, services.WebRTC, log),
	}
	
	// Инициализируем анонимные handlers если сервисы доступны
	if services.AnonymousRoom != nil {
		handlers.AnonymousRoom = NewAnonymousRoomHandler(services.AnonymousRoom, log)
		log.Info("AnonymousRoom handler initialized")
	}
	if services.AnonymousMedia != nil {
		handlers.AnonymousMedia = NewAnonymousMediaHandler(services.AnonymousMedia, log)
		log.Info("AnonymousMedia handler initialized")
	}
	
	// Инициализируем анонимный чат
	if repos.AnonymousChat != nil && repos.AnonymousRoom != nil {
		handlers.AnonymousChat = NewAnonymousChatHandler(repos.AnonymousChat, repos.AnonymousRoom, log)
		log.Info("AnonymousChat handler initialized")
	}
	
	return handlers
}
