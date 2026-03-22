package service

import (
	"fmt"

	"telegram-service/logger"
)

// ApplyPromoCodeFromMessage применяет промокод пользователю через команду /promo CODE.
// Использует SendText для отправки результата (так как вызывается из Message, а не CallbackQuery).
// Логирует успешную активацию в логгер.
func (s *service) ApplyPromoCodeFromMessage(chatID int64, promoCode string, logger *logger.MyLogger) error {
	// 1. Получить промокод (проверка валидности, лимита, истечения)
	bonusDays, promoCodeID, err := s.postgres.GetPromoCode(promoCode)
	if err != nil {
		msg := fmt.Sprintf("❌ Промокод '%s' недействителен", promoCode)
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("invalid promo code: %w", err)
	}

	// 2. Проверить, может ли пользователь активировать этот код
	canActivate, err := s.postgres.CanActivatePromoCode(promoCodeID, chatID)
	if err != nil {
		msg := "❌ Ошибка при проверке промокода"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}
	if !canActivate {
		msg := "❌ Вы уже активировали этот промокод"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}

	// 3. Проверить, есть ли у пользователя пир с ключами
	hasPeer, err := s.postgres.HasPeerWithKeys(chatID)
	if err != nil {
		msg := "❌ Ошибка при проверке конфигурации"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}
	if !hasPeer {
		msg := "❌ У вас нет сконфигурированного пира. Сначала используйте бесплатный тариф или купите подписку, затем активируйте промокод"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}

	// 4. Добавить бонусные дни к подписке
	if err := s.postgres.ApplyPromoBonusDays(chatID, bonusDays); err != nil {
		msg := "❌ Ошибка при применении бонусных дней"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}

	// 5. Восстановить пира в WireGuard (если он был истекший)
	peer, err := s.postgres.GetConnection(chatID)
	if err == nil && peer.PublicKey != "" && peer.PresharedKey != "" {
		socket := fmt.Sprintf("10.66.66.%d/32", peer.HostID)
		if err := s.httpClient.RestorePeer(peer.PublicKey, peer.PresharedKey, socket, chatID); err != nil {
			msg := "❌ Ошибка при восстановлении пера в WireGuard"
			s.telegram.SendText(chatID, msg)
			return fmt.Errorf("promo code activation failed")
		}
	}

	// 6. Записать активацию промокода
	if err := s.postgres.ActivatePromoCode(promoCodeID, chatID); err != nil {
		msg := "❌ Ошибка при активации промокода"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}

	// 7. Увеличить счетчик использований промокода
	if err := s.postgres.IncrementPromoCodeUsage(promoCodeID); err != nil {
		msg := "❌ Ошибка при обновлении счетчика использований"
		s.telegram.SendText(chatID, msg)
		return fmt.Errorf("promo code activation failed")
	}

	// 8. Отправить успешное сообщение пользователю
	s.telegram.SendText(chatID, fmt.Sprintf("✅ Промокод успешно активирован! Добавлено +%d дней к вашей подписке", bonusDays))
	logger.Logger.Info(fmt.Sprintf("пользователь %v применил промокод %s", chatID, promoCode))
	return nil
}
