package service

import (
	"context"
	"fmt"
	"time"

	"telegram-service/logger"
)

const text = "Ваша подписка истекла, продлите для дальнейшего использования нашей услуги"

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
			}
			for _, r := range data {
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
			logger.Logger.Info("проверка подписок завершена")
		}
	}
}
