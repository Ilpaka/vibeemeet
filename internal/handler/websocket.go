package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"video_conference/internal/service"
	"video_conference/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // В продакшене нужно проверять origin
	},
}

type WebSocketHandler struct {
	chatService service.ChatService
	log         logger.Logger
}

func NewWebSocketHandler(chatService service.ChatService, log logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		chatService: chatService,
		log:         log,
	}
}

func (h *WebSocketHandler) HandleChat(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("Failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	_ = roomID // TODO: использовать roomID для валидации

	// TODO: реализовать обработку WebSocket сообщений для чата
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			h.log.Error("Failed to read message", "error", err)
			break
		}

		// Echo обратно
		if err := conn.WriteMessage(messageType, message); err != nil {
			h.log.Error("Failed to write message", "error", err)
			break
		}
	}
}

