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

	hostID, err := s.postgres.GetHostID(chatID)
	if err != nil {
		return fmt.Errorf("Error getting host ID: %w", err)
	}

	// Add peer to AWG and get both keys in one call
	data, err := s.httpClient.AddPeer(hostID, true, chatID)
	if err != nil {
		return fmt.Errorf("Error adding peer: %w", err)
	}

	publicKey := data.GetKey()
	presharedKey := data.GetPresharedKey()

	// Save connection with all data in one database call
	if err := s.postgres.SaveConnection(chatID, publicKey, presharedKey, expiresAt); err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	// Download and send config file
	bufer, err := s.httpClient.DownloadConfFile(chatID)
	if err != nil {
		return fmt.Errorf("Error downloading config file: %w", err)
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
