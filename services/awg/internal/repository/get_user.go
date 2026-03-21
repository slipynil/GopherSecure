package repository

import (
	"awg-service/internal/repository/model"
	"encoding/json"
	"fmt"
	"os"
)

// GetUser retrieves a user by ID from the users.json file.
func (r *Repository) GetUser(id int64) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var users []model.User
	content, err := os.ReadFile(r.UsersFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read users file: %w", err)
	}

	if err := json.Unmarshal(content, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	for _, user := range users {
		if user.Id == id {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user with id %d not found", id)
}
