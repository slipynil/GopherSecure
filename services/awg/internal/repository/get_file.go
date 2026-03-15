package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

func (r *Repository) GetFile(id string) (string, error) {
	filePath := filepath.Join(r.ConfDirPath, id+".conf")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not exist")
	} else if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}
	return filePath, nil
}
