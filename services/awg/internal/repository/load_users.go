package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	if err = json.Unmarshal(data, &users); err != nil {
		return err
	}

	// Получаем список уже существующих пиров
	out, err := exec.Command("awg", "show", r.Device, "peers").Output()
	if err != nil {
		return fmt.Errorf("failed to list peers: %w", err)
	}
	existingPeers := string(out)

	for _, user := range users {
		// Пропускаем если пир уже существует
		if strings.Contains(existingPeers, user.PublicKey) {
			continue
		}

		// Preshared key передаётся через временный файл
		tmpFile, err := os.CreateTemp("", "psk-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}

		_, err = tmpFile.WriteString(user.PresharedKey)
		tmpFile.Close()
		if err != nil {
			os.Remove(tmpFile.Name())
			return fmt.Errorf("failed to write preshared key: %w", err)
		}

		err = exec.Command(
			"awg", "set", r.Device,
			"peer", user.PublicKey,
			"preshared-key", tmpFile.Name(),
			"allowed-ips", user.VirtualEndpoint,
		).Run()

		os.Remove(tmpFile.Name())

		if err != nil {
			return fmt.Errorf("failed to restore peer %s: %w", user.PublicKey, err)
		}
	}
	return nil
}
