package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"video_conference/internal/domain"
	"video_conference/pkg/logger"
)

const (
	// TTL для чата - 6 часов
	ChatTTL = 6 * time.Hour
	
	// Префикс ключей Redis
	ChatMessagesKeyPrefix = "chat:room:%s:messages"
	RoomParticipantsKeyPrefix = "room:%s:participants"
	RoomMetaKeyPrefix = "room:%s:meta"
)

type AnonymousChatRepository interface {
	// Сохранить сообщение в Redis (TTL 6 часов)
	SaveMessage(ctx context.Context, roomID uuid.UUID, message *domain.AnonymousChatMessage) error
	
	// Получить сообщения из Redis (последние N, отсортированные по времени)
	GetMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.AnonymousChatMessage, error)
	
	// Получить сообщения после определенного времени
	GetMessagesAfter(ctx context.Context, roomID uuid.UUID, after time.Time, limit int) ([]*domain.AnonymousChatMessage, error)
	
	// Удалить сообщение
	DeleteMessage(ctx context.Context, roomID uuid.UUID, messageID string) error
	
	// Обновить сообщение
	UpdateMessage(ctx context.Context, roomID uuid.UUID, message *domain.AnonymousChatMessage) error
	
	// Проверить существование комнаты в Redis
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type anonymousChatRepository struct {
	rdb *redis.Client
	log logger.Logger
}

func NewAnonymousChatRepository(rdb *redis.Client, log logger.Logger) AnonymousChatRepository {
	return &anonymousChatRepository{
		rdb: rdb,
		log: log,
	}
}

func (r *anonymousChatRepository) getMessagesKey(roomID uuid.UUID) string {
	return fmt.Sprintf(ChatMessagesKeyPrefix, roomID.String())
}

func (r *anonymousChatRepository) SaveMessage(ctx context.Context, roomID uuid.UUID, message *domain.AnonymousChatMessage) error {
	key := r.getMessagesKey(roomID)
	
	// Сериализуем сообщение в JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		r.log.Error("Failed to marshal message", "error", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Используем timestamp в миллисекундах как score для сортировки
	score := float64(message.CreatedAt.UnixMilli())
	
	// Добавляем в sorted set
	err = r.rdb.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: messageJSON,
	}).Err()
	if err != nil {
		r.log.Error("Failed to save message to Redis", "error", err, "room_id", roomID)
		return fmt.Errorf("failed to save message: %w", err)
	}
	
	// Устанавливаем TTL на ключ (6 часов)
	err = r.rdb.Expire(ctx, key, ChatTTL).Err()
	if err != nil {
		r.log.Warn("Failed to set TTL on chat key", "error", err)
		// Не критичная ошибка, продолжаем
	}
	
	return nil
}

func (r *anonymousChatRepository) GetMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.AnonymousChatMessage, error) {
	key := r.getMessagesKey(roomID)
	
	// Получаем последние N сообщений (от новых к старым)
	// Используем ZREVRANGE для получения в обратном порядке
	messagesJSON, err := r.rdb.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []*domain.AnonymousChatMessage{}, nil
		}
		r.log.Error("Failed to get messages from Redis", "error", err, "room_id", roomID)
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	
	messages := make([]*domain.AnonymousChatMessage, 0, len(messagesJSON))
	for _, msgJSON := range messagesJSON {
		var message domain.AnonymousChatMessage
		if err := json.Unmarshal([]byte(msgJSON), &message); err != nil {
			r.log.Warn("Failed to unmarshal message", "error", err)
			continue
		}
		messages = append(messages, &message)
	}
	
	// Разворачиваем массив, чтобы получить хронологический порядок (от старых к новым)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}

func (r *anonymousChatRepository) GetMessagesAfter(ctx context.Context, roomID uuid.UUID, after time.Time, limit int) ([]*domain.AnonymousChatMessage, error) {
	key := r.getMessagesKey(roomID)
	
	// Минимальный score (timestamp в миллисекундах)
	minScore := float64(after.UnixMilli())
	
	// Получаем сообщения после указанного времени
	messagesJSON, err := r.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%.0f", minScore),
		Max:    "+inf",
		Offset: 0,
		Count:  int64(limit),
	}).Result()
	
	if err != nil {
		if err == redis.Nil {
			return []*domain.AnonymousChatMessage{}, nil
		}
		r.log.Error("Failed to get messages after time", "error", err, "room_id", roomID)
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	
	messages := make([]*domain.AnonymousChatMessage, 0, len(messagesJSON))
	for _, msgJSON := range messagesJSON {
		var message domain.AnonymousChatMessage
		if err := json.Unmarshal([]byte(msgJSON), &message); err != nil {
			r.log.Warn("Failed to unmarshal message", "error", err)
			continue
		}
		messages = append(messages, &message)
	}
	
	return messages, nil
}

func (r *anonymousChatRepository) DeleteMessage(ctx context.Context, roomID uuid.UUID, messageID string) error {
	key := r.getMessagesKey(roomID)
	
	// Получаем все сообщения и находим нужное
	messagesJSON, err := r.rdb.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("message not found")
		}
		r.log.Error("Failed to get messages for deletion", "error", err)
		return fmt.Errorf("failed to get messages: %w", err)
	}
	
	// Ищем сообщение с нужным ID
	for _, msgJSON := range messagesJSON {
		var message domain.AnonymousChatMessage
		if err := json.Unmarshal([]byte(msgJSON), &message); err != nil {
			continue
		}
		
		if message.ID == messageID {
			// Удаляем из sorted set
			err = r.rdb.ZRem(ctx, key, msgJSON).Err()
			if err != nil {
				r.log.Error("Failed to delete message from Redis", "error", err)
				return fmt.Errorf("failed to delete message: %w", err)
			}
			return nil
		}
	}
	
	return errors.New("message not found")
}

func (r *anonymousChatRepository) UpdateMessage(ctx context.Context, roomID uuid.UUID, message *domain.AnonymousChatMessage) error {
	key := r.getMessagesKey(roomID)
	
	// Получаем все сообщения и находим нужное
	messagesJSON, err := r.rdb.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("message not found")
		}
		r.log.Error("Failed to get messages for update", "error", err)
		return fmt.Errorf("failed to get messages: %w", err)
	}
	
	// Ищем сообщение с нужным ID
	for _, msgJSON := range messagesJSON {
		var oldMessage domain.AnonymousChatMessage
		if err := json.Unmarshal([]byte(msgJSON), &oldMessage); err != nil {
			continue
		}
		
		if oldMessage.ID == message.ID {
			// Удаляем старое сообщение
			err = r.rdb.ZRem(ctx, key, msgJSON).Err()
			if err != nil {
				r.log.Error("Failed to remove old message", "error", err)
				return fmt.Errorf("failed to update message: %w", err)
			}
			
			// Добавляем обновленное сообщение
			newMessageJSON, err := json.Marshal(message)
			if err != nil {
				r.log.Error("Failed to marshal updated message", "error", err)
				return fmt.Errorf("failed to marshal message: %w", err)
			}
			
			score := float64(message.CreatedAt.UnixMilli())
			err = r.rdb.ZAdd(ctx, key, redis.Z{
				Score:  score,
				Member: newMessageJSON,
			}).Err()
			if err != nil {
				r.log.Error("Failed to add updated message", "error", err)
				return fmt.Errorf("failed to update message: %w", err)
			}
			
			// Обновляем TTL
			r.rdb.Expire(ctx, key, ChatTTL)
			
			return nil
		}
	}
	
	return errors.New("message not found")
}

func (r *anonymousChatRepository) RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	key := r.getMessagesKey(roomID)
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		r.log.Error("Failed to check room existence", "error", err)
		return false, err
	}
	return exists > 0, nil
}

