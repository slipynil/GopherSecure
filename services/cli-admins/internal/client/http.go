package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PromoClient отправляет HTTP запросы к API телеграм сервиса для управления промокодами.
type PromoClient struct {
	baseURL string
	client  *http.Client
}

// NewPromoClient создает новый клиент с указанным базовым URL.
func NewPromoClient(baseURL string) *PromoClient {
	return &PromoClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// CreatePromoRequest структура для создания промокода.
type CreatePromoRequest struct {
	Code      string `json:"code"`
	BonusDays int    `json:"bonus_days"`
	MaxUses   int    `json:"max_uses"`
	ExpiresAt string `json:"expires_at"` // формат: 2026-03-29T23:59:59Z
}

// UpdatePromoRequest структура для обновления промокода.
type UpdatePromoRequest struct {
	BonusDays int    `json:"bonus_days"`
	MaxUses   int    `json:"max_uses"`
	ExpiresAt string `json:"expires_at"`
}

// CreatePromo создает новый промокод.
func (c *PromoClient) CreatePromo(req CreatePromoRequest) (map[string]any, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		fmt.Sprintf("http://%s/admin/promo", c.baseURL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp, http.StatusCreated)
}

// UpdatePromo обновляет параметры промокода.
func (c *PromoClient) UpdatePromo(promoID int, req UpdatePromoRequest) (map[string]any, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/admin/promo/%d", c.baseURL, promoID),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp, http.StatusOK)
}

// ListPromos получает список всех промокодов.
func (c *PromoClient) ListPromos() ([]map[string]any, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://%s/admin/promo", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// DeletePromo удаляет промокод (мягкое удаление).
func (c *PromoClient) DeletePromo(promoID int) (map[string]any, error) {
	httpReq, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/admin/promo/%d", c.baseURL, promoID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp, http.StatusOK)
}

// parseResponse парсит JSON ответ от сервера.
func (c *PromoClient) parseResponse(resp *http.Response, expectedStatus int) (map[string]any, error) {
	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
