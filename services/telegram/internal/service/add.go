package service

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// add добавляет новое подключение WireGuard для пользователя.
// Если price равен 20000 копеек, создает подписку на 30 дней, иначе на 24 часа.
// Функция создает запись в БД, добавляет пира на AWG сервисе и отправляет конфиг файл пользователю.
// В случае ошибки на любом этапе возвращает описание проблемы.
func (s *service) add(chatID int64, price int) error {
	now := time.Now()
	duration := 2 * time.Minute
	if price == 20000 {
		duration = 5 * time.Minute
	}
	expiresAt := now.Add(duration)

	// Check if peer already exists with keys (renewal case)
	existing, err := s.postgres.GetConnection(chatID)
	if err == nil && existing.PublicKey != "" {
		// Renewal: restore existing peer in WireGuard with the same keys
		hostID := existing.HostID
		socket := fmt.Sprintf("10.66.66.%d/32", hostID)
		if err := s.httpClient.RestorePeer(existing.PublicKey, existing.PresharedKey, socket); err != nil {
			return fmt.Errorf("failed to restore peer: %w", err)
		}
		if err := s.postgres.SaveConnection(hostID, existing.PublicKey, existing.PresharedKey, expiresAt); err != nil {
			return fmt.Errorf("failed to update connection expiry: %w", err)
		}
	} else {
		// New peer: create placeholder, add to AWG, save keys
		hostID, err := s.postgres.NewConnection(chatID)
		if err != nil {
			return fmt.Errorf("failed to create connection: %w", err)
		}
		data, err := s.httpClient.AddPeer(hostID, true, chatID)
		if err != nil {
			return fmt.Errorf("failed to add peer: %w", err)
		}
		if err := s.postgres.SaveConnection(hostID, data.GetKey(), data.GetPresharedKey(), expiresAt); err != nil {
			return fmt.Errorf("failed to save connection: %w", err)
		}
	}

	// Download and send config file
	bufer, err := s.httpClient.DownloadConfFile(chatID)
	if err != nil {
		return fmt.Errorf("failed to download config file: %w", err)
	}
	return s.telegram.SendFile(chatID, bufer)
}

// getConfFile отправляет файл конфигурации пользователю.
// Предварительно проверяет, имеет ли пользователь активную подписку.
// Если подписка отсутствует, отправляет уведомление об этом.
func (s *service) getConfFile(u tgbotapi.Update) error {
	chatID := u.CallbackQuery.Message.Chat.ID
	status, err := s.postgres.CheckStatus(chatID)
	if err != nil {
		s.telegram.UpdateSendText(u, "Ошибка проверки статуса")
		return fmt.Errorf("failed to check status: %w", err)
	}
	if !status {
		s.telegram.UpdateSendText(u, "у вас нет подписки")
		return nil
	}
	// get http response buffer of config file
	bufer, err := s.httpClient.DownloadConfFile(chatID)
	if err != nil {
		return fmt.Errorf("Error downloading config file: %w", err)
	}
	return s.telegram.SendFile(chatID, bufer)
}
