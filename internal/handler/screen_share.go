package handler

import (
	"context"
	_ "embed"
	"net/http"
	"strings"
	"sync"

	"video_conference/internal/service"
	"video_conference/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v4"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//go:embed screen_share.html
var htmlPage string

type ScreenShareHandler struct {
	screenCapture   service.ScreenCaptureService
	audioCapture    service.AudioCaptureService
	webrtcService   service.WebRTCService
	log             logger.Logger
	peerConnections map[uuid.UUID]*webrtc.PeerConnection
	iceCandidates   map[uuid.UUID][]webrtc.ICECandidateInit
	mu              sync.RWMutex
}

func NewScreenShareHandler(
	screenCapture service.ScreenCaptureService,
	audioCapture service.AudioCaptureService,
	webrtcService service.WebRTCService,
	log logger.Logger,
) *ScreenShareHandler {
	return &ScreenShareHandler{
		screenCapture:   screenCapture,
		audioCapture:    audioCapture,
		webrtcService:   webrtcService,
		log:             log,
		peerConnections: make(map[uuid.UUID]*webrtc.PeerConnection),
		iceCandidates:   make(map[uuid.UUID][]webrtc.ICECandidateInit),
	}
}

type OfferRequest struct {
	SDP  string `json:"sdp" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type AnswerResponse struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// HandleOffer обрабатывает WebRTC offer от клиента
func (h *ScreenShareHandler) HandleOffer(c *gin.Context) {
	h.log.Info("Received offer request")

	var req OfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to parse request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Parsed offer request", "type", req.Type, "sdp_length", len(req.SDP))

	// Создаем контекст для захвата - НЕ используем c.Request.Context(),
	// так как он отменяется при завершении HTTP запроса
	// Используем context.Background() чтобы контекст не отменялся автоматически
	ctx := context.Background()

	// Запускаем захват экрана (может включать звук через GetDisplayMedia)
	h.log.Info("Starting screen capture")
	screenStream, err := h.screenCapture.StartCapture(ctx)
	if err != nil {
		h.log.Error("Failed to start screen capture", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start screen capture: " + err.Error()})
		return
	}
	h.log.Info("Screen capture started successfully")

	// Проверяем, есть ли уже аудио в screenStream (если использовался GetDisplayMedia)
	audioTracks := screenStream.GetAudioTracks()
	var audioStream mediadevices.MediaStream

	if len(audioTracks) > 0 {
		// GetDisplayMedia уже включил звук, используем его
		h.log.Info("Audio already included in screen stream via GetDisplayMedia", "audio_tracks", len(audioTracks))
		audioStream = screenStream // Используем тот же стрим
	} else {
		// GetDisplayMedia не включил звук, пробуем отдельный захват
		h.log.Info("No audio in screen stream, starting separate audio capture")
		audioStream, err = h.audioCapture.StartCapture(ctx)
		if err != nil {
			h.log.Warn("Failed to start audio capture, continuing without audio", "error", err)
			audioStream = nil // Продолжаем без звука
		} else {
			h.log.Info("Separate audio capture started successfully")
		}
	}

	// Парсим тип SDP
	var sdpType webrtc.SDPType
	switch req.Type {
	case "offer":
		sdpType = webrtc.SDPTypeOffer
	case "answer":
		sdpType = webrtc.SDPTypeAnswer
	default:
		h.log.Error("Invalid SDP type", "type", req.Type)
		for _, track := range screenStream.GetTracks() {
			track.Close()
		}
		if audioStream != nil {
			for _, track := range audioStream.GetTracks() {
				track.Close()
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid SDP type"})
		return
	}

	// Создаем PeerConnection (без треков пока)
	h.log.Info("Creating peer connection")
	peerConnection, err := h.webrtcService.CreatePeerConnection(ctx, screenStream, audioStream)
	if err != nil {
		h.log.Error("Failed to create peer connection", "error", err)
		for _, track := range screenStream.GetTracks() {
			track.Close()
		}
		if audioStream != nil {
			for _, track := range audioStream.GetTracks() {
				track.Close()
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create peer connection: " + err.Error()})
		return
	}
	h.log.Info("Peer connection created successfully")

	// Сохраняем соединение
	peerConnectionID := uuid.New()
	h.mu.Lock()
	h.peerConnections[peerConnectionID] = peerConnection
	h.mu.Unlock()

	// Обработка ICE candidates от сервера - сохраняем их для отправки клиенту
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			h.log.Info("Server ICE candidate generated", "candidate", candidate.String())
			h.mu.Lock()
			if h.iceCandidates[peerConnectionID] == nil {
				h.iceCandidates[peerConnectionID] = make([]webrtc.ICECandidateInit, 0)
			}
			h.iceCandidates[peerConnectionID] = append(h.iceCandidates[peerConnectionID], candidate.ToJSON())
			h.mu.Unlock()
		} else {
			h.log.Info("Server ICE candidate gathering complete")
		}
	})

	// ВАЖНО: Добавляем треки ДО установки remote description!
	// Это правильный порядок для WebRTC - треки должны быть добавлены перед установкой remote description
	h.log.Info("Adding tracks to peer connection BEFORE setting remote description")
	if err := h.webrtcService.AddTracksToPeerConnection(peerConnection, screenStream, audioStream); err != nil {
		h.log.Error("Failed to add tracks", "error", err)
		h.cleanup(peerConnectionID, screenStream, audioStream)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add tracks: " + err.Error()})
		return
	}
	h.log.Info("Tracks added successfully")

	// Устанавливаем remote description (offer от клиента) ПОСЛЕ добавления треков
	offer := webrtc.SessionDescription{
		Type: sdpType,
		SDP:  req.SDP,
	}

	h.log.Info("Setting remote description", "type", req.Type, "sdp_length", len(req.SDP), "has_ice_ufrag", strings.Contains(req.SDP, "ice-ufrag"))

	// Логируем кодеки из offer для диагностики
	offerLines := strings.Split(req.SDP, "\n")
	var offerVideoCodecs []string
	inVideoSection := false
	for _, line := range offerLines {
		if strings.HasPrefix(line, "m=video") {
			inVideoSection = true
			parts := strings.Fields(line)
			if len(parts) > 3 {
				h.log.Info("Offer video media line", "codecs", strings.Join(parts[3:], " "))
			}
		} else if strings.HasPrefix(line, "m=") {
			inVideoSection = false
		} else if inVideoSection && strings.HasPrefix(line, "a=rtpmap:") {
			if strings.Contains(line, "VP8") || strings.Contains(line, "VP9") || strings.Contains(line, "H264") {
				offerVideoCodecs = append(offerVideoCodecs, line)
			}
		}
	}
	if len(offerVideoCodecs) > 0 {
		h.log.Info("Offer video codecs", "codecs", strings.Join(offerVideoCodecs, "; "))
	}
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		h.log.Error("Failed to set remote description", "error", err)
		h.cleanup(peerConnectionID, screenStream, audioStream)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set remote description: " + err.Error()})
		return
	}
	h.log.Info("Remote description set successfully")

	// Создаем answer
	h.log.Info("Creating answer")
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		h.log.Error("Failed to create answer", "error", err)
		h.cleanup(peerConnectionID, screenStream, audioStream)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create answer: " + err.Error()})
		return
	}

	// Проверяем answer.SDP перед установкой local description
	h.log.Info("Checking answer SDP", "answer_sdp_length", len(answer.SDP), "has_ice_ufrag", strings.Contains(answer.SDP, "ice-ufrag"))

	// Логируем информацию о кодеках в SDP для диагностики
	hasVP8 := strings.Contains(answer.SDP, "VP8")
	hasVP9 := strings.Contains(answer.SDP, "VP9")
	hasH264 := strings.Contains(answer.SDP, "H264") || strings.Contains(answer.SDP, "H.264")

	h.log.Info("Answer SDP codec information",
		"has_vp8", hasVP8,
		"has_vp9", hasVP9,
		"has_h264", hasH264)

	// Извлекаем информацию о видео кодеках из SDP
	lines := strings.Split(answer.SDP, "\n")
	var videoCodecs []string
	inVideoSection = false
	for _, line := range lines {
		if strings.HasPrefix(line, "m=video") {
			inVideoSection = true
			// Извлекаем кодеки из m=video строки
			parts := strings.Fields(line)
			if len(parts) > 3 {
				h.log.Info("Video media line", "codecs", strings.Join(parts[3:], " "))
			}
		} else if strings.HasPrefix(line, "m=") {
			inVideoSection = false
		} else if inVideoSection && strings.HasPrefix(line, "a=rtpmap:") {
			if strings.Contains(line, "VP8") || strings.Contains(line, "VP9") || strings.Contains(line, "H264") {
				videoCodecs = append(videoCodecs, line)
			}
		}
	}

	if len(videoCodecs) > 0 {
		h.log.Info("Video codecs in SDP", "codecs", strings.Join(videoCodecs, "; "))
	} else {
		h.log.Warn("No video codecs found in SDP answer!")
	}

	// Логируем первые 1000 символов SDP для отладки
	if len(answer.SDP) > 1000 {
		h.log.Info("Answer SDP preview", "sdp_preview", answer.SDP[:1000])
	} else {
		h.log.Info("Answer SDP", "sdp", answer.SDP)
	}

	// Устанавливаем local description ПЕРЕД отправкой answer
	h.log.Info("Setting local description")
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		h.log.Error("Failed to set local description", "error", err)
		h.cleanup(peerConnectionID, screenStream, audioStream)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set local description: " + err.Error()})
		return
	}

	// Получаем обновленный SDP после установки local description
	localDescription := peerConnection.LocalDescription()
	var sdp string
	if localDescription != nil {
		sdp = localDescription.SDP
		h.log.Info("Got local description", "sdp_length", len(sdp), "has_ice_ufrag", strings.Contains(sdp, "ice-ufrag"))
	} else {
		// Fallback на answer.SDP если localDescription nil
		h.log.Warn("Local description is nil, using answer.SDP")
		sdp = answer.SDP
	}

	// Проверяем, что SDP содержит ICE credentials
	if !strings.Contains(sdp, "ice-ufrag") || !strings.Contains(sdp, "ice-pwd") {
		h.log.Error("SDP does not contain ICE credentials", "sdp_preview", sdp[:min(500, len(sdp))])
		// Попробуем использовать answer.SDP
		if strings.Contains(answer.SDP, "ice-ufrag") && strings.Contains(answer.SDP, "ice-pwd") {
			h.log.Info("Using answer.SDP as fallback")
			sdp = answer.SDP
		} else {
			h.log.Error("Answer SDP also does not contain ICE credentials")
		}
	}

	// Отправляем answer клиенту вместе с ID соединения
	h.log.Info("Sending answer to client", "peer_connection_id", peerConnectionID.String(), "sdp_length", len(sdp), "has_ice_ufrag", strings.Contains(sdp, "ice-ufrag"), "has_ice_pwd", strings.Contains(sdp, "ice-pwd"))
	c.Header("X-Peer-Connection-ID", peerConnectionID.String())
	c.JSON(http.StatusOK, AnswerResponse{
		SDP:  sdp,
		Type: "answer",
	})
}

// HandleICE обрабатывает ICE candidates от клиента
func (h *ScreenShareHandler) HandleICE(c *gin.Context) {
	peerConnectionIDStr := c.Param("id")
	peerConnectionID, err := uuid.Parse(peerConnectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer connection ID"})
		return
	}

	h.mu.RLock()
	peerConnection, exists := h.peerConnections[peerConnectionID]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "peer connection not found"})
		return
	}

	var candidate webrtc.ICECandidateInit
	if err := c.ShouldBindJSON(&candidate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := peerConnection.AddICECandidate(candidate); err != nil {
		h.log.Error("Failed to add ICE candidate", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add ICE candidate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetICE возвращает ICE candidates от сервера
func (h *ScreenShareHandler) GetICE(c *gin.Context) {
	peerConnectionIDStr := c.Param("id")
	peerConnectionID, err := uuid.Parse(peerConnectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer connection ID"})
		return
	}

	h.mu.RLock()
	candidates, exists := h.iceCandidates[peerConnectionID]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusOK, gin.H{"candidates": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}

// HandleHangup закрывает соединение
func (h *ScreenShareHandler) HandleHangup(c *gin.Context) {
	peerConnectionIDStr := c.Param("id")
	peerConnectionID, err := uuid.Parse(peerConnectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer connection ID"})
		return
	}

	h.cleanup(peerConnectionID, nil, nil)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ServeHTML отдает HTML страницу для клиента
func (h *ScreenShareHandler) ServeHTML(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlPage)
}

func (h *ScreenShareHandler) cleanup(peerConnectionID uuid.UUID, screenStream, audioStream mediadevices.MediaStream) {
	h.mu.Lock()
	peerConnection, exists := h.peerConnections[peerConnectionID]
	if exists {
		delete(h.peerConnections, peerConnectionID)
	}
	delete(h.iceCandidates, peerConnectionID)
	h.mu.Unlock()

	if exists && peerConnection != nil {
		peerConnection.Close()
	}

	if screenStream != nil {
		for _, track := range screenStream.GetTracks() {
			track.Close()
		}
	}

	if audioStream != nil {
		for _, track := range audioStream.GetTracks() {
			track.Close()
		}
	}

	h.screenCapture.StopCapture()
	if audioStream != nil {
		h.audioCapture.StopCapture()
	}
}
