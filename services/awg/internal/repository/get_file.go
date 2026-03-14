package repository

import (
	"fmt"
	"os"
	"path"
)

func (r *Repository) GetFile(id string) (string, error) {
	filePath := path.Join(r.StoragePath, id+".conf")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not exist")
	} else if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}
	return filePath, nil
}
