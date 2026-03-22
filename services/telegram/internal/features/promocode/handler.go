package promocode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"telegram-service/logger"

	"github.com/labstack/echo/v5"
)

// Handler содержит зависимости для обработки HTTP запросов промокодов.
type Handler struct {
	postgres postgres
	logger   *logger.MyLogger
}

// postgres определяет интерфейс доступа к БД для операций с промокодами.
type postgres interface {
	CreatePromoCode(code string, bonusDays, maxUses int, expiresAt time.Time) (int, error)
	UpdatePromoCode(promoCodeID, bonusDays, maxUses int, expiresAt time.Time) error
	GetAllPromoCodes() ([]map[string]any, error)
	DeactivatePromoCode(promoCodeID int) error
}

// NewHandler создает новый handler для управления промокодами.
func NewHandler(postgres postgres, logger *logger.MyLogger) *Handler {
	return &Handler{
		postgres: postgres,
		logger:   logger,
	}
}

// CreatePromo обрабатывает POST /admin/promo — создание нового промокода.
func (h *Handler) CreatePromo(c *echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.logger.IsErr("failed to read body", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	var req CreatePromoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.IsErr("failed to parse request", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	// Валидация
	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "code is required"})
	}
	if req.BonusDays <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bonus_days must be > 0"})
	}
	if req.MaxUses < 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "max_uses must be >= 0"})
	}

	// Парсить дату
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		h.logger.IsErr("failed to parse expires_at", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid expires_at format (use RFC3339)"})
	}

	// Создать промокод
	promoID, err := h.postgres.CreatePromoCode(req.Code, req.BonusDays, req.MaxUses, expiresAt)
	if err != nil {
		// Если это ошибка дубликата — некритичная ошибка
		if strings.Contains(err.Error(), "duplicate key") {
			msg := fmt.Sprintf("промокод '%s' уже был создан", req.Code)
			h.logger.Logger.Info(msg)
			return c.JSON(http.StatusConflict, ErrorResponse{Error: msg})
		}
		h.logger.IsErr("failed to create promo code", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create promo code"})
	}

	msg := "создан новый промокод " + req.Code
	h.logger.Logger.Info(msg)

	return c.JSON(http.StatusCreated, SuccessResponse{
		Message: "promo code created successfully",
		Data: map[string]any{
			"id":         promoID,
			"code":       req.Code,
			"bonus_days": req.BonusDays,
			"max_uses":   req.MaxUses,
			"expires_at": expiresAt,
		},
	})
}

// UpdatePromo обрабатывает PUT /admin/promo/:id — обновление параметров промокода.
func (h *Handler) UpdatePromo(c *echo.Context) error {
	idStr := c.Param("id")
	promoID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid promo id"})
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.logger.IsErr("failed to read body", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	var req UpdatePromoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.IsErr("failed to parse request", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	// Валидация
	if req.BonusDays <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bonus_days must be > 0"})
	}
	if req.MaxUses < 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "max_uses must be >= 0"})
	}

	// Парсить дату
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		h.logger.IsErr("failed to parse expires_at", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid expires_at format (use RFC3339)"})
	}

	// Обновить промокод
	err = h.postgres.UpdatePromoCode(promoID, req.BonusDays, req.MaxUses, expiresAt)
	if err != nil {
		h.logger.IsErr("failed to update promo code", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update promo code"})
	}

	msg := "промокод #" + idStr + " обновлен"
	h.logger.Logger.Info(msg)

	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "promo code updated successfully",
		Data: map[string]any{
			"id":         promoID,
			"bonus_days": req.BonusDays,
			"max_uses":   req.MaxUses,
			"expires_at": expiresAt,
		},
	})
}

// ListPromos обрабатывает GET /admin/promo — получение списка всех промокодов.
func (h *Handler) ListPromos(c *echo.Context) error {
	promos, err := h.postgres.GetAllPromoCodes()
	if err != nil {
		h.logger.IsErr("failed to get all promo codes", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get promo codes"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": promos,
	})
}

// DeletePromo обрабатывает DELETE /admin/promo/:id — удаление (деактивация) промокода.
func (h *Handler) DeletePromo(c *echo.Context) error {
	idStr := c.Param("id")
	promoID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid promo id"})
	}

	// Деактивировать промокод
	err = h.postgres.DeactivatePromoCode(promoID)
	if err != nil {
		h.logger.IsErr("failed to deactivate promo code", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete promo code"})
	}

	msg := "промокод #" + idStr + " деактивирован"
	h.logger.Logger.Info(msg)

	return c.JSON(http.StatusOK, SuccessResponse{
		Message: "promo code deleted successfully",
	})
}

// RegisterRoutes регистрирует маршруты для управления промокодами.
func RegisterRoutes(e *echo.Echo, handler *Handler) {
	admin := e.Group("/admin")
	admin.POST("/promo", handler.CreatePromo)
	admin.PUT("/promo/:id", handler.UpdatePromo)
	admin.GET("/promo", handler.ListPromos)
	admin.DELETE("/promo/:id", handler.DeletePromo)
}
