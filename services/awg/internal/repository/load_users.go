package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"awg-service/internal/repository/model"
)

func (r *Repository) LoadUsers() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.UsersFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}

	var users []model.User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return err
	}

	for _, user := range users {
		err := exec.Command(
			"awg", "set", r.Device,
			"peer", user.PublicKey,
			"preshared-key", user.PresharedKey,
			"allowed-ips", user.VirtualEndpoint,
		).Run()
		if err != nil {
			return fmt.Errorf("failed to restore peer %s: %w", user.PublicKey, err)
		}
	}
	return nil
}
