package promocode

import "time"

// CreatePromoRequest структура запроса для создания промокода.
type CreatePromoRequest struct {
	Code      string `json:"code"`
	BonusDays int    `json:"bonus_days"`
	MaxUses   int    `json:"max_uses"`
	ExpiresAt string `json:"expires_at"` // RFC3339 format
}

// UpdatePromoRequest структура запроса для обновления промокода.
type UpdatePromoRequest struct {
	BonusDays int    `json:"bonus_days"`
	MaxUses   int    `json:"max_uses"`
	ExpiresAt string `json:"expires_at"`
}

// PromoCodeResponse структура ответа с информацией о промокоде.
type PromoCodeResponse struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	BonusDays int       `json:"bonus_days"`
	MaxUses   int       `json:"max_uses"`
	UsedCount int       `json:"used_count"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SuccessResponse структура успешного ответа.
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse структура ошибочного ответа.
type ErrorResponse struct {
	Error string `json:"error"`
}
