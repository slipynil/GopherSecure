package main

import (
	getenv "awg-service/internal/getEnv"
	"awg-service/internal/repository"
	"awg-service/internal/transport"
	"awg-service/internal/transport/handlers"
	"awg-service/logger"
	"fmt"
	"os"
	"path/filepath"

	awgctrlgo "github.com/slipynil/awgctrl-go"
)

var (
	DefaultHTTP   = ""
	DefaultDEVICE = ""
	DefaultAWG    = ""
)

func main() {
	logger := logger.New()
	cfg, err := getenv.NewObfuscation()

	if err != nil {
		logger.Fatal(err)
	}

	tunnelName, awgEndpoint := getOpt(os.Getenv("DEVICE"), DefaultDEVICE), getOpt(os.Getenv("AWG_ENDPOINT"), DefaultAWG)
	httpEndpoint := getOpt(os.Getenv("HTTP_ENDPOINT"), DefaultHTTP)

	if tunnelName == "" || awgEndpoint == "" || httpEndpoint == "" {
		err := fmt.Errorf("DEVICE and AWG_ENDPOINT environment variables are required")
		logger.Fatal(err)
	}

	storagePath, err := filepath.Abs("/etc/amnezia/amneziawg/configs/")
	if err != nil {
		logger.Fatal(err)
	}

	awg, err := awgctrlgo.New(tunnelName, awgEndpoint, storagePath, cfg)
	if err != nil {
		logger.Fatal(err)
	}
	repository := repository.New(storagePath)
	handlers := handlers.New(awg, repository)
	service := transport.New(handlers)

	service.Start(httpEndpoint)
}

func getOpt(value string, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
