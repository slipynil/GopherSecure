package service

import (
	"context"
	"fmt"
	"time"

	"telegram-service/logger"
)

// text содержит сообщение об истечении подписки.
const text = "Ваша подписка истекла, продлите для дальнейшего использования нашей услуги"

// CheckSubcription периодически проверяет истекшие подписки с заданным интервалом duration.
// Для каждой истекшей подписки удаляет пира на AWG сервисе, обновляет статус в БД и отправляет уведомление пользователю.
// Функция работает до отмены контекста ctx.
func (s *service) CheckSubcription(ctx context.Context, logger *logger.MyLogger, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data, err := s.postgres.ExpiredConnection()
			if err != nil {
				logger.IsErr("fail to get expired connections", err)
				continue
			}

			if len(data) > 0 {
				logger.Logger.Info(fmt.Sprintf("found %d expired connections", len(data)))
			}

			for _, r := range data {
				if r.PublicKey == "" {
					logger.Logger.Info(fmt.Sprintf("skipping peer deletion - PublicKey is empty for ChatID: %d", r.ChatID))
					continue
				}

				if err := s.httpClient.DeletePeer(r.PublicKey); err != nil {
					logger.IsErr("fail to delete peer", err)
				}
				if err := s.postgres.StatusFalse(r.ChatID); err != nil {
					logger.IsErr("fail to update status", err)
				}
				if err = s.telegram.SendText(r.ChatID, text); err != nil {
					logger.IsErr("fail to send text", err)
				}
				msg := fmt.Sprintf("у пользователя %d закончилась подписка", r.ChatID)
				logger.Logger.Info(msg)
			}
		}
	}
}
