package main

import (
	"context"
	"os"
	httpclient "telegram-service/internal/httpClient"
	"telegram-service/internal/repository"
	"telegram-service/internal/service"
	"telegram-service/internal/telegram"
	"telegram-service/logger"
)

func main() {
	tgToken := os.Getenv("TELEGRAM_KEY")
	providerToken := os.Getenv("PROVIDER_TOKEN")
	url := os.Getenv("HTTP_URL")
	dbConn := os.Getenv("DB_CONN")

	if len(tgToken) == 0 || len(providerToken) == 0 || len(url) == 0 || len(dbConn) == 0 {
		panic("TELEGRAM_KEY, PROVIDER_TOKEN, HTTP_URL, or DB_CONN environment variable is not set")
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
	logger, closeLogger, _ := logger.NewLogger()
	defer closeLogger()

	// run service with background context
	ctx := context.Background()
	service.Update(ctx, logger)
}
