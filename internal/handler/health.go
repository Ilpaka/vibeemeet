package handler

import (
	"net/http"

	"video_conference/internal/config"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	hostIP      string
	liveKitPort string
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	hostIP := cfg.LiveKit.HostIP
	if hostIP == "" {
		hostIP = config.GetLocalIP()
	}

	// Используем порт из конфига или дефолтный
	liveKitPort := cfg.LiveKit.Port
	if liveKitPort == "" {
		liveKitPort = "7880"
	}

	return &HealthHandler{
		hostIP:      hostIP,
		liveKitPort: liveKitPort,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "video-conference",
	})
}

// ServerInfo возвращает информацию о сервере для клиентов
func (h *HealthHandler) ServerInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"host_ip":       h.hostIP,
		"livekit_port":  h.liveKitPort,
		"livekit_url":   "ws://" + h.hostIP + ":" + h.liveKitPort,
		"api_base":      "/api/v1",
	})
}

