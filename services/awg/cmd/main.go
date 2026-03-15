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

	tunnelName := getOpt(os.Getenv("DEVICE"), DefaultDEVICE)
	awgEndpoint := getOpt(os.Getenv("AWG_ENDPOINT"), DefaultAWG)
	httpEndpoint := getOpt(os.Getenv("HTTP_ENDPOINT"), DefaultHTTP)

	if tunnelName == "" ||
		awgEndpoint == "" ||
		httpEndpoint == "" {
		err := fmt.Errorf("DEVICE and AWG_ENDPOINT environment variables are required")
		logger.Fatal(err)
	}

	parrentDirPath, err := filepath.Abs("/etc/amnezia/amneziawg/")
	if err != nil {
		logger.Fatal(err)
	}
	repository := repository.New(parrentDirPath, tunnelName)
	if err := repository.LoadUsers(); err != nil {
		logger.Fatal(err)
	}

	awg, err := awgctrlgo.New(tunnelName, awgEndpoint, repository.ConfDirPath, cfg)
	if err != nil {
		logger.Fatal(err)
	}
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
