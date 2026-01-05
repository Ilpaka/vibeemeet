package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnonymousRoom - упрощенная модель комнаты без привязки к пользователю
type AnonymousRoom struct {
	ID              uuid.UUID  `json:"id"`
	LiveKitRoomName string     `json:"livekit_room_name"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	Status          string     `json:"status"` // active, ended
	MaxParticipants int        `json:"max_participants"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"` // Автоматическое удаление неактивных комнат
}

// AnonymousParticipant - анонимный участник комнаты
type AnonymousParticipant struct {
	ID            uuid.UUID  `json:"id"`
	RoomID        uuid.UUID  `json:"room_id"`
	ParticipantID string     `json:"participant_id"` // Временный ID с клиента (UUID строка)
	DisplayName   string     `json:"display_name"`
	Role          string     `json:"role"` // host, participant
	LiveKitSID    *string    `json:"livekit_sid,omitempty"`
	JoinedAt      time.Time  `json:"joined_at"`
	LeftAt        *time.Time `json:"left_at,omitempty"`
	ClientIP      *string    `json:"client_ip,omitempty"`
	UserAgent     *string    `json:"user_agent,omitempty"`
}

// AnonymousChatMessage - сообщение чата для Redis
type AnonymousChatMessage struct {
	ID            string    `json:"id"`            // UUID строка
	RoomID        uuid.UUID `json:"room_id"`       // UUID комнаты
	ParticipantID string    `json:"participant_id"` // Временный ID участника
	DisplayName   string    `json:"display_name"`  // Имя отправителя
	MessageType   string    `json:"message_type"`   // user, system
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
}

// Константы для типов сообщений
const (
	AnonymousMessageTypeUser   = "user"
	AnonymousMessageTypeSystem = "system"
)

// Примечание: Константы RoomStatusActive, RoomStatusEnded, ParticipantRoleHost, 
// ParticipantRoleParticipant уже определены в room.go и используются здесь

