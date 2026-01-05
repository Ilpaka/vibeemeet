package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"video_conference/internal/service"
	"video_conference/pkg/logger"
)

type WaitingRoomHandler struct {
	roomService service.RoomService
	log         logger.Logger
}

func NewWaitingRoomHandler(roomService service.RoomService, log logger.Logger) *WaitingRoomHandler {
	return &WaitingRoomHandler{
		roomService: roomService,
		log:         log,
	}
}

func (h *WaitingRoomHandler) List(c *gin.Context) {
	// TODO: реализовать получение списка ожидающих
	c.JSON(http.StatusOK, gin.H{"message": "Not implemented"})
}

func (h *WaitingRoomHandler) Approve(c *gin.Context) {
	// TODO: реализовать одобрение входа
	entryID, _ := uuid.Parse(c.Param("entryId"))
	_ = entryID
	c.JSON(http.StatusOK, gin.H{"message": "Not implemented"})
}

func (h *WaitingRoomHandler) Reject(c *gin.Context) {
	// TODO: реализовать отклонение входа
	entryID, _ := uuid.Parse(c.Param("entryId"))
	_ = entryID
	c.JSON(http.StatusOK, gin.H{"message": "Not implemented"})
}

