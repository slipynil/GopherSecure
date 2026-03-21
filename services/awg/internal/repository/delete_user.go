package repository

import (
	"awg-service/internal/repository/model"
	"awg-service/internal/transport/dto"
	"encoding/json"
	"os"
	"slices"
)

func (r *Repository) DeleteUser(publicKey string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.UsersFilePath)
	if err != nil {
		return err
	}

	var users []model.User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return err
	}

	index := slices.IndexFunc(users, func(user model.User) bool {
		return user.PublicKey == publicKey
	})
	if index == -1 {
		return nil
	}
	users = slices.Delete(users, index, index+1)

	file, err := os.OpenFile(r.UsersFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(users)
}

// DeleteUserEx deletes a user and returns detailed result information.
// Safe for rollback: returns the deleted user so it can be restored if needed.
func (r *Repository) DeleteUserEx(publicKey string) (*dto.DeleteResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.UsersFilePath)
	if err != nil {
		return nil, err
	}

	var users []model.User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(users, func(user model.User) bool {
		return user.PublicKey == publicKey
	})
	if index == -1 {
		// User not found
		return &dto.DeleteResult{
			Found:   false,
			Deleted: false,
			User:    nil,
		}, nil
	}

	// Save user for potential rollback
	deletedUser := users[index]
	users = slices.Delete(users, index, index+1)

	file, err := os.OpenFile(r.UsersFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(users)
	if err != nil {
		return nil, err
	}

	return &dto.DeleteResult{
		Found:   true,
		Deleted: true,
		User:    &deletedUser,
	}, nil
}

// RestoreUser restores a deleted user back to users.json.
// Used for rollback when AWG deletion fails.
func (r *Repository) RestoreUser(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.UsersFilePath)
	if err != nil {
		return err
	}

	var users []model.User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return err
	}

	// Check if user already exists (shouldn't happen, but be safe)
	existingIndex := slices.IndexFunc(users, func(u model.User) bool {
		return u.PublicKey == user.PublicKey
	})
	if existingIndex != -1 {
		// User already exists, no need to restore
		return nil
	}

	// Add user back
	users = append(users, *user)

	file, err := os.OpenFile(r.UsersFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(users)
}
