package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParticipantMiddleware проверяет наличие participant_id в заголовке X-Participant-ID
// Если нет - генерирует новый UUID и добавляет в контекст
func ParticipantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		participantID := c.GetHeader("X-Participant-ID")
		
		// Валидация UUID формата
		if participantID != "" {
			if _, err := uuid.Parse(participantID); err != nil {
				// Невалидный UUID, генерируем новый
				participantID = ""
			}
		}
		
		// Если нет валидного participant_id, генерируем новый
		if participantID == "" {
			participantID = uuid.New().String()
		}
		
		// Сохраняем в контекст
		c.Set("participant_id", participantID)
		
		// Добавляем в заголовок ответа для клиента
		c.Header("X-Participant-ID", participantID)
		
		c.Next()
	}
}

