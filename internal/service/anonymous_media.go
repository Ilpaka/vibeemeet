package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"video_conference/internal/config"
	"video_conference/internal/domain"
	"video_conference/internal/repository"
	"video_conference/pkg/logger"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
)

type AnonymousMediaService interface {
	GetToken(ctx context.Context, roomID uuid.UUID, participantID string, displayName string) (string, string, error)
}

type anonymousMediaService struct {
	roomRepo repository.AnonymousRoomRepository
	cfg      config.LiveKitConfig
	log      logger.Logger
}

func NewAnonymousMediaService(roomRepo repository.AnonymousRoomRepository, cfg config.LiveKitConfig, log logger.Logger) AnonymousMediaService {
	return &anonymousMediaService{
		roomRepo: roomRepo,
		cfg:      cfg,
		log:      log,
	}
}

func (s *anonymousMediaService) GetToken(ctx context.Context, roomID uuid.UUID, participantID string, displayName string) (string, string, error) {
	// Проверяем существование комнаты
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		// Если комнаты нет, создаем её автоматически (временное решение для тестирования)
		s.log.Info("Room not found, creating automatically", "room_id", roomID, "participant_id", participantID)
		now := time.Now()
		room = &domain.AnonymousRoom{
			ID:              roomID,
			LiveKitRoomName: roomID.String(),
			Title:           "Auto-created room",
			Status:          domain.RoomStatusActive,
			MaxParticipants: 10,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if createErr := s.roomRepo.Create(ctx, room); createErr != nil {
			s.log.Error("Failed to auto-create room", "error", createErr)
			return "", "", errors.New("room not found and failed to create")
		}
	}

	// Проверяем, что комната активна (используем константу из domain)
	if room.Status != domain.RoomStatusActive {
		return "", "", errors.New("room is not active")
	}

	// Создаем токен доступа LiveKit
	at := auth.NewAccessToken(s.cfg.APIKey, s.cfg.APISecret)
	canPublish := true
	canSubscribe := true
	grant := &auth.VideoGrant{
		RoomJoin:     true,
		Room:         room.LiveKitRoomName,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}

	// Используем participant_id как identity (должен быть валидным UUID)
	// Если participantID не валидный UUID, используем его как есть (LiveKit поддерживает строки)
	identity := participantID
	if _, err := uuid.Parse(participantID); err != nil {
		// Если не UUID, генерируем UUID на основе participantID для consistency
		identity = uuid.NewSHA1(uuid.NameSpaceOID, []byte(participantID)).String()
	}

	at.AddGrant(grant).
		SetIdentity(identity).
		SetName(displayName).
		SetValidFor(time.Hour)

	token, err := at.ToJWT()
	if err != nil {
		s.log.Error("Failed to generate LiveKit token", "error", err, "room_id", roomID, "participant_id", participantID)
		return "", "", errors.New("failed to generate token")
	}

	// Формируем URL для фронтенда
	url := s.buildFrontendURL()

	s.log.Info("LiveKit token generated",
		"room_id", roomID,
		"participant_id", participantID,
		"url", url,
		"host_ip", s.cfg.HostIP,
	)

	return token, url, nil
}

// buildFrontendURL формирует правильный URL для подключения с фронтенда
func (s *anonymousMediaService) buildFrontendURL() string {
	// Если явно указан FrontendURL, используем его
	if s.cfg.FrontendURL != "" {
		url := s.cfg.FrontendURL
		// Заменяем localhost на HostIP если указан
		if s.cfg.HostIP != "" && s.cfg.HostIP != "localhost" {
			url = strings.Replace(url, "localhost", s.cfg.HostIP, 1)
		}
		// Заменяем Docker hostname
		if strings.Contains(url, "livekit:") {
			url = strings.Replace(url, "livekit:", s.cfg.HostIP+":", 1)
		}
		return s.normalizeWSURL(url)
	}

	// Формируем URL на основе HostIP
	hostIP := s.cfg.HostIP
	if hostIP == "" {
		hostIP = "localhost"
	}

	// LiveKit по умолчанию работает на порту 7880
	url := "ws://" + hostIP + ":7880"

	return url
}

// normalizeWSURL убеждается, что URL имеет правильный формат WebSocket
func (s *anonymousMediaService) normalizeWSURL(url string) string {
	// Убираем trailing slash
	url = strings.TrimSuffix(url, "/")

	// Преобразуем http/https в ws/wss
	if strings.HasPrefix(url, "https://") {
		url = "wss://" + strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		url = "ws://" + strings.TrimPrefix(url, "http://")
	}

	// Добавляем протокол если отсутствует
	if !strings.HasPrefix(url, "ws://") && !strings.HasPrefix(url, "wss://") {
		url = "ws://" + url
	}

	return url
}

