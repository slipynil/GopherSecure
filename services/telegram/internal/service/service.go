// Package service предоставляет основную бизнес-логику для управления VPN подписками и взаимодействием с клиентами через Telegram.
package service

import (
	"context"
	"fmt"
	"time"

	"telegram-service/internal/dto"
	"telegram-service/logger"
)

// Update запускает основной цикл обработки обновлений от Telegram.
// Функция обрабатывает сообщения, платежи и взаимодействие пользователя с кнопками,
// а также периодически проверяет истекшие подписки через [CheckSubcription].
func (s *service) Update(ctx context.Context, logger *logger.MyLogger) {

	logger.Logger.Info("service is running")
	duration := time.Hour
	go s.CheckSubcription(ctx, logger, duration)

	for u := range s.telegram.Chan() {
		// если есть сигнал об оплате
		if u.PreCheckoutQuery != nil {
			err := s.telegram.PreCheckoutQuery(u)
			logger.IsErr("failed to answer pre_checkout_query", err)
			// если структура message не пустая и является командой
		} else if u.Message != nil {
			chat := u.Message.Chat
			if u.Message.SuccessfulPayment != nil {
				if dto, err := s.telegram.HandleSuccessfulPayment(u); err != nil {
					logger.IsErr("payment canceled", err)
				} else {
					err := s.postgres.SuccessfulPaymentStatus(dto.InvoicePayload)
					logger.IsErr("failed to change true status on payment", err)
					err = s.postgres.StatusTrue(chat.ID)
					logger.IsErr("failed to change true status on client", err)
					err = s.add(chat.ID, dto.TotalAmount)
					logger.IsErr("failed to add peer", err)
					if err == nil {
						msg := fmt.Sprintf("пользователь %s купил подписку за %v %s", chat.UserName, dto.TotalAmount/100, dto.Currency)
						logger.Logger.Info(msg)
					}
				}
			}
			// если команда
			switch u.Message.Command() {

			case "menu":
				s.telegram.Menu(chat.ID)

			case "start":
				err := s.postgres.AddClient(chat.UserName, chat.ID)
				if err != nil {
					logger.IsErr("failed to add client to postgres", err)
				} else {
					err := s.telegram.SendText(chat.ID, "Вы успешно авторизовались, нажмите /menu для вывода всех функций")
					logger.IsErr("", err)
				}
			}
			continue
			// если это inline кнопка
		} else if u.CallbackQuery != nil {

			chat := u.CallbackQuery.Message.Chat

			// add peer command send conf file for user
			callBackData, err := dto.DecodeCallbackData(u.CallbackQuery.Data)
			logger.IsErr("failed decoding callback data", err)

			switch callBackData.Action {

			case "<- назад":
				err := s.telegram.UpdateMainMenu(u)
				logger.IsErr("failed to redirect main menu", err)

			case "получить конфиг":
				err := s.getConfFile(u)
				logger.IsErr("failed to get conf file", err)

			case "помощь":
				err := s.telegram.UpdateSendText(u, HelpText)
				logger.IsErr("", err)

			case "стоимость":
				err := s.telegram.UpdateSendText(u, PricingText)
				logger.IsErr("", err)

			case "оплатить":
				status, err := s.postgres.CheckStatus(chat.ID)
				if err != nil {
					logger.IsErr("failed to check status", err)
					err := s.telegram.UpdateSendText(u, "Ошибка проверки статуса")
					logger.IsErr("", err)
				} else if status {
					err := s.telegram.UpdateSendText(u, "Вы уже оплатили")
					logger.IsErr("", err)
				} else {
					err := s.Invoice(u)
					if err != nil {
						logger.IsErr("failed to create invoice", err)
					}
				}

			case "протестировать":
				isTested, err := s.postgres.IsTested(chat.ID)
				if err != nil {
					logger.IsErr("failed to check if tested", err)
					err := s.telegram.UpdateSendText(u, "Ошибка проверки статуса")
					logger.IsErr("", err)
				} else if isTested {
					err := s.telegram.UpdateSendText(u, "У вас уже был тестовый доступ")
					logger.IsErr("", err)
				} else {
					err := s.postgres.Tested(chat.ID)
					logger.IsErr("failed to mark user as tested", err)
					err = s.add(chat.ID, 0)
					logger.IsErr("failed to add user", err)
					err = s.telegram.UpdateSendText(u, "Тестовый доступ активирован на 24 часа")
					if err != nil {
						logger.IsErr("failed to send text", err)
					} else {
						msg := fmt.Sprintf("пользователь %s получил тестовый доступ", chat.UserName)
						logger.Logger.Info(msg)
					}
				}
			}
		}
	}
}
