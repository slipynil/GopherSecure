package repository

import (
	"awg-service/internal/repository/model"
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
