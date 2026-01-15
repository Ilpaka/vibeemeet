package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AuthServiceClient представляет клиент для обращения к Auth-сервису
type AuthServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAuthServiceClient создает новый клиент для Auth-сервиса
func NewAuthServiceClient(baseURL string) *AuthServiceClient {
	return &AuthServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GuestCreateRequest - запрос на создание гостевой сессии
type GuestCreateRequest struct {
	DisplayName string  `json:"display_name"`
	RoomID      *string `json:"room_id,omitempty"`
}

// GuestCreateResponse - ответ с гостевой сессией
type GuestCreateResponse struct {
	GuestID    string    `json:"guest_id"`
	GuestToken string    `json:"guest_token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// VerifyTokenRequest - запрос на проверку токена
type VerifyTokenRequest struct {
	Token string `json:"token"`
}

// VerifyTokenResponse - ответ с информацией о токене
type VerifyTokenResponse struct {
	Valid       bool     `json:"valid"`
	UserID      *string  `json:"user_id,omitempty"`
	Email       *string  `json:"email,omitempty"`
	DisplayName *string  `json:"display_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	IsGuest     bool     `json:"is_guest"`
	ExpiresAt   *int64   `json:"expires_at,omitempty"`
}

// CreateGuestSession создает гостевую сессию на Auth-сервисе
func (c *AuthServiceClient) CreateGuestSession(ctx context.Context, displayName, roomID string) (*GuestCreateResponse, error) {
	req := GuestCreateRequest{
		DisplayName: displayName,
		RoomID:      &roomID,
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/guest", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var response GuestCreateResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}

// VerifyToken проверяет токен на Auth-сервисе
func (c *AuthServiceClient) VerifyToken(ctx context.Context, token string) (*VerifyTokenResponse, error) {
	req := VerifyTokenRequest{
		Token: token,
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/verify", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var response VerifyTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}

// VerifyGuestToken проверяет гостевой токен
func (c *AuthServiceClient) VerifyGuestToken(ctx context.Context, token string) (*VerifyTokenResponse, error) {
	req := VerifyTokenRequest{
		Token: token,
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/guest/verify", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var response VerifyTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}
