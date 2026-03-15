package service

import (
	"errors"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-service/internal/dto"
)

func (s *service) add(chatID int64, price int) error {
	expiresAt := s.calculateExpiration(price)

	_, _, err := s.postgres.GetPeer(chatID)
	if err == nil {
		return s.renewPeer(chatID, expiresAt)
	}
	if !errors.Is(err, dto.ErrNotFound) {
		return fmt.Errorf("failed to check existing peer: %w", err)
	}

	// DB first — generates host_id for correct virtual IP (10.66.66.N)
	hostID, err := s.postgres.NewConnection(chatID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	data, err := s.httpClient.AddPeer(hostID, true, chatID)
	if err != nil {
		_ = s.postgres.DeleteConnection(chatID)
		return fmt.Errorf("failed to add peer to AWG: %w", err)
	}

	if err := s.postgres.SaveKeys(chatID, data.GetKey(), data.GetPresharedKey()); err != nil {
		_ = s.httpClient.DeletePeer(data.GetKey())
		_ = s.postgres.DeleteConnection(chatID)
		return fmt.Errorf("failed to save keys: %w", err)
	}

	return s.sendConfigToUser(chatID)
}

func (s *service) calculateExpiration(price int) time.Time {
	duration := 24 * time.Hour
	if price == 20000 {
		duration = 30 * 24 * time.Hour
	}
	return time.Now().Add(duration)
}

func (s *service) renewPeer(chatID int64, expiresAt time.Time) error {
	if err := s.postgres.RenewConnection(chatID, expiresAt); err != nil {
		return fmt.Errorf("failed to renew subscription: %w", err)
	}
	return s.sendConfigToUser(chatID)
}

func (s *service) sendConfigToUser(chatID int64) error {
	buf, err := s.httpClient.DownloadConfFile(chatID)
	if err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}
	return s.telegram.SendFile(chatID, buf)
}

func (s *service) getConfFile(u tgbotapi.Update) error {
	chatID := u.CallbackQuery.Message.Chat.ID
	if !s.postgres.CheckStatus(chatID) {
		s.telegram.UpdateSendText(u, "у вас нет подписки")
		return nil
	}
	buf, err := s.httpClient.DownloadConfFile(chatID)
	if err != nil {
		return fmt.Errorf("failed to download config file: %w", err)
	}
	return s.telegram.SendFile(chatID, buf)
}
