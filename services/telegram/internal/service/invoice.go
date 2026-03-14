package service

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/google/uuid"
)

func (s *service) Invoice(u tgbotapi.Update) error {
	chatID := u.CallbackQuery.Message.Chat.ID
	payload := uuid.New().String()

	// create new entity payment in postgres
	if err := s.postgres.NewPayment(chatID, payload); err != nil {
		return fmt.Errorf("fail to create new payment entity in postgres %w", err)
	}
	// create new invoice in telegram
	if err := s.telegram.CreateAndSendInvoice(chatID, payload); err != nil {
		return fmt.Errorf("fail to create and send invoice for telegram client")
	}
	return nil
}
