package main

import (
	"context"
	"log/slog"
	"os"

	"telegram-service/internal/features/promocode"
	httpclient "telegram-service/internal/httpClient"
	"telegram-service/internal/repository"
	"telegram-service/internal/service"
	"telegram-service/internal/telegram"
	"telegram-service/logger"

	"github.com/labstack/echo/v5"
)

func main() {
	tgToken := os.Getenv("TELEGRAM_KEY")
	providerToken := os.Getenv("PROVIDER_TOKEN")
	url := os.Getenv("HTTP_URL")
	dbConn := os.Getenv("DB_CONN")
	adminAddress := os.Getenv("ADMIN_ADDRESS")

	if len(tgToken) == 0 || len(providerToken) == 0 || len(url) == 0 || len(dbConn) == 0 {
		panic("TELEGRAM_KEY, PROVIDER_TOKEN, HTTP_URL, or DB_CONN environment variable is not set")
	}

	if len(adminAddress) == 0 {
		adminAddress = "0.0.0.0:8080"
	}

	// init telegram service
	telegram, err := telegram.New(tgToken, providerToken)
	if err != nil {
		panic(err)
	}

	// init http client service
	httpClient := httpclient.New(url)

	// init postgres service
	postgres, err := repository.New(context.Background(), dbConn)
	if err != nil {
		panic(err)
	}
	defer postgres.Close()

	// init service
	service := service.New(telegram, httpClient, postgres)

	// init logger
	myLogger, closeLogger, _ := logger.NewLogger()
	defer closeLogger()

	// init echo http server for admin API
	e := echo.New()
	e.Logger = slog.New(NewZapHandler(myLogger))

	// register promo code routes
	promoHandler := promocode.NewHandler(postgres, myLogger)
	promocode.RegisterRoutes(e, promoHandler)

	// start HTTP server in background
	go func() {
		myLogger.Logger.Info("admin API server started on " + adminAddress)
		if err := e.Start(adminAddress); err != nil {
			myLogger.Logger.Fatal("failed to start admin API server: " + err.Error())
		}
	}()

	// run telegram bot service
	ctx := context.Background()
	service.Update(ctx, myLogger)
}

// ZapHandler адаптирует zap.Logger для slog интерфейса
type ZapHandler struct {
	myLogger *logger.MyLogger
}

// NewZapHandler создает адаптер
func NewZapHandler(myLogger *logger.MyLogger) *ZapHandler {
	return &ZapHandler{myLogger: myLogger}
}

// Handle обрабатывает slog запись
func (h *ZapHandler) Handle(ctx context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		h.myLogger.Logger.Debug(r.Message)
	case slog.LevelInfo:
		h.myLogger.Logger.Info(r.Message)
	case slog.LevelWarn:
		h.myLogger.Logger.Warn(r.Message)
	case slog.LevelError:
		h.myLogger.Logger.Error(r.Message)
	}
	return nil
}

// WithAttrs добавляет атрибуты (требуется slog.Handler интерфейсом)
func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup добавляет группу (требуется slog.Handler интерфейсом)
func (h *ZapHandler) WithGroup(name string) slog.Handler {
	return h
}

// Enabled проверяет уровень логирования (требуется slog.Handler интерфейсом)
func (h *ZapHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}
