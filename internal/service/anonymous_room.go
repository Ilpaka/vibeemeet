package service

import (
	"context"
	"errors"
	"time"

	"video_conference/internal/config"
	"video_conference/internal/domain"
	"video_conference/internal/repository"
	"video_conference/pkg/logger"

	"github.com/google/uuid"
)

type AnonymousRoomService interface {
	Create(ctx context.Context, title string, description *string, maxParticipants int, participantID string, displayName string) (*domain.AnonymousRoom, *domain.AnonymousParticipant, error)
	GetByID(ctx context.Context, roomID uuid.UUID) (*domain.AnonymousRoom, error)
	Join(ctx context.Context, roomID uuid.UUID, participantID string, displayName string) (*domain.AnonymousParticipant, error)
	Leave(ctx context.Context, roomID uuid.UUID, participantID string) error
	GetParticipants(ctx context.Context, roomID uuid.UUID) ([]*domain.AnonymousParticipant, error)
}

type anonymousRoomService struct {
	roomRepo repository.AnonymousRoomRepository
	cfg      *config.Config
	log      logger.Logger
}

func NewAnonymousRoomService(roomRepo repository.AnonymousRoomRepository, cfg *config.Config, log logger.Logger) AnonymousRoomService {
	return &anonymousRoomService{
		roomRepo: roomRepo,
		cfg:      cfg,
		log:      log,
	}
}

func (s *anonymousRoomService) Create(ctx context.Context, title string, description *string, maxParticipants int, participantID string, displayName string) (*domain.AnonymousRoom, *domain.AnonymousParticipant, error) {
	if maxParticipants <= 0 || maxParticipants > 500 {
		maxParticipants = 10
	}

	if title == "" {
		title = "Новая комната"
	}

	// Создаем комнату
	roomID := uuid.New()
	now := time.Now()

	room := &domain.AnonymousRoom{
		ID:              roomID,
		LiveKitRoomName: roomID.String(),
		Title:           title,
		Description:     description,
		Status:          domain.RoomStatusActive,
		MaxParticipants: maxParticipants,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		s.log.Error("Failed to create anonymous room", "error", err)
		return nil, nil, errors.New("failed to create room")
	}

	// Создаем участника (host)
	participant := &domain.AnonymousParticipant{
		ID:            uuid.New(),
		RoomID:        roomID,
		ParticipantID: participantID,
		DisplayName:   displayName,
		Role:          domain.ParticipantRoleHost,
		JoinedAt:      now,
	}

	if err := s.roomRepo.CreateParticipant(ctx, participant); err != nil {
		s.log.Error("Failed to create participant", "error", err)
		// Не критично, продолжаем
	}

	s.log.Info("Anonymous room created", "room_id", roomID, "participant_id", participantID)

	return room, participant, nil
}

func (s *anonymousRoomService) GetByID(ctx context.Context, roomID uuid.UUID) (*domain.AnonymousRoom, error) {
	return s.roomRepo.GetByID(ctx, roomID)
}

func (s *anonymousRoomService) Join(ctx context.Context, roomID uuid.UUID, participantID string, displayName string) (*domain.AnonymousParticipant, error) {
	// Проверяем существование комнаты
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, errors.New("room not found")
	}

	if room.Status != domain.RoomStatusActive {
		return nil, errors.New("room is not active")
	}

	// Проверяем, не присоединился ли уже этот участник
	existingParticipant, err := s.roomRepo.GetParticipant(ctx, roomID, participantID)
	if err == nil && existingParticipant != nil && existingParticipant.LeftAt == nil {
		// Участник уже в комнате
		return existingParticipant, nil
	}

	// Создаем нового участника
	participant := &domain.AnonymousParticipant{
		ID:            uuid.New(),
		RoomID:        roomID,
		ParticipantID: participantID,
		DisplayName:   displayName,
		Role:          domain.ParticipantRoleParticipant,
		JoinedAt:      time.Now(),
	}

	if err := s.roomRepo.CreateParticipant(ctx, participant); err != nil {
		s.log.Error("Failed to create participant", "error", err)
		return nil, errors.New("failed to join room")
	}

	return participant, nil
}

func (s *anonymousRoomService) Leave(ctx context.Context, roomID uuid.UUID, participantID string) error {
	// Отмечаем, что участник вышел
	if err := s.roomRepo.MarkParticipantLeft(ctx, roomID, participantID); err != nil {
		s.log.Error("Failed to mark participant left", "error", err, "room_id", roomID, "participant_id", participantID)
		return err
	}

	// Проверяем количество активных участников
	count, err := s.roomRepo.GetActiveParticipantCount(ctx, roomID)
	if err != nil {
		s.log.Error("Failed to get active participant count", "error", err, "room_id", roomID)
		return nil // Не блокируем выход ошибкой подсчета
	}

	// Если участников нет, деактивируем комнату
	if count == 0 {
		s.log.Info("Room is empty, deactivating", "room_id", roomID)
		if err := s.roomRepo.SetRoomStatus(ctx, roomID, domain.RoomStatusEnded); err != nil {
			s.log.Error("Failed to deactivate empty room", "error", err, "room_id", roomID)
		}
	}

	return nil
}

func (s *anonymousRoomService) GetParticipants(ctx context.Context, roomID uuid.UUID) ([]*domain.AnonymousParticipant, error) {
	return s.roomRepo.GetParticipantsByRoom(ctx, roomID)
}
