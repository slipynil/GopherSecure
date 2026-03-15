package repository

import (
	"os"
	"path/filepath"
	"sync"
)

type Repository struct {
	GlobalFilePath string
	ConfDirPath    string
	Device         string
	mu             sync.Mutex
}

func New(dirPath string, device string) *Repository {
	globalFilePath := filepath.Join(dirPath, "data", "users.json")
	confDirPath := filepath.Join(dirPath, "configures")
	os.Mkdir(confDirPath, 0755)
	os.MkdirAll(filepath.Join(dirPath, "data"), 0755)
	return &Repository{
		GlobalFilePath: globalFilePath,
		ConfDirPath:    confDirPath,
		Device:         device,
	}
}
