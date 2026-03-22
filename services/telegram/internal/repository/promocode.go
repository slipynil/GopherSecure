package repository

import (
	"time"
)

// GetPromoCode возвращает информацию о промокоде по коду.
// Проверяет, что код активен, не истек и не исчерпан лимит использований.
func (p *Postgres) GetPromoCode(code string) (bonusDays int, promoCodeID int, err error) {
	sqlRaw := `
	SELECT id, bonus_days
	FROM promo_codes
	WHERE code = $1
		AND is_active = TRUE
		AND expires_at > NOW()
		AND (max_uses = 0 OR used_count < max_uses)
	LIMIT 1;
	`
	err = p.conn.QueryRow(p.ctx, sqlRaw, code).Scan(&promoCodeID, &bonusDays)
	if err != nil {
		return 0, 0, err
	}
	return bonusDays, promoCodeID, nil
}

// CanActivatePromoCode проверяет, может ли пользователь активировать промокод.
// Возвращает true если пользователь еще не активировал этот код.
func (p *Postgres) CanActivatePromoCode(promoCodeID int, chatID int64) (bool, error) {
	sqlRaw := `
	SELECT 1 FROM promo_activations
	WHERE promo_code_id = $1 AND chat_id = $2
	LIMIT 1;
	`
	var exists int
	err := p.conn.QueryRow(p.ctx, sqlRaw, promoCodeID, chatID).Scan(&exists)

	if err != nil && err.Error() != "no rows in result set" {
		return false, err
	}

	// Если ошибка "no rows" — пользователь еще не активировал, вернуть true
	if err != nil {
		return true, nil
	}

	// Если существует запись — уже активировал, вернуть false
	return false, nil
}

// ActivatePromoCode создает запись об активации промокода пользователем.
func (p *Postgres) ActivatePromoCode(promoCodeID int, chatID int64) error {
	sqlRaw := `
	INSERT INTO promo_activations (promo_code_id, chat_id)
	VALUES ($1, $2);
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, promoCodeID, chatID)
	return err
}

// IncrementPromoCodeUsage увеличивает счетчик использований промокода на 1.
func (p *Postgres) IncrementPromoCodeUsage(promoCodeID int) error {
	sqlRaw := `
	UPDATE promo_codes
	SET used_count = used_count + 1
	WHERE id = $1;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, promoCodeID)
	return err
}

// ApplyPromoBonusDays добавляет бонусные дни к подписке пользователя.
// Увеличивает expires_at текущего пира на bonusDays.
// Обновляет пир независимо от текущего статуса (даже если пир истекший).
func (p *Postgres) ApplyPromoBonusDays(chatID int64, bonusDays int) error {
	sqlRaw := `
	UPDATE peer
	SET expires_at = CASE
		WHEN expires_at > NOW() THEN expires_at + INTERVAL '1 day' * $1
		ELSE NOW() + INTERVAL '1 day' * $1
	END
	WHERE chat_id = $2;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, bonusDays, chatID)
	return err
}


// DeactivatePromoCode устанавливает is_active = FALSE для промокода (мягкое удаление).
func (p *Postgres) DeactivatePromoCode(promoCodeID int) error {
	sqlRaw := `
	UPDATE promo_codes
	SET is_active = FALSE
	WHERE id = $1;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, promoCodeID)
	return err
}

// GetPromoCodeStats возвращает статистику по промокоду: количество использований и оставшийся лимит.
func (p *Postgres) GetPromoCodeStats(promoCodeID int) (used, maxUses int, expiresAt time.Time, err error) {
	sqlRaw := `
	SELECT used_count, max_uses, expires_at
	FROM promo_codes
	WHERE id = $1;
	`
	err = p.conn.QueryRow(p.ctx, sqlRaw, promoCodeID).Scan(&used, &maxUses, &expiresAt)
	return used, maxUses, expiresAt, err
}

// CreatePromoCode создает новый промокод.
// bonusDays — количество дней бонуса (например, 30, 60, 90).
// maxUses — максимальное количество использований (0 = безлимит).
// expiresAt — когда истекает действие кода.
func (p *Postgres) CreatePromoCode(code string, bonusDays, maxUses int, expiresAt time.Time) (int, error) {
	sqlRaw := `
	INSERT INTO promo_codes (code, bonus_days, max_uses, expires_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
	`
	var promoCodeID int
	err := p.conn.QueryRow(p.ctx, sqlRaw, code, bonusDays, maxUses, expiresAt).Scan(&promoCodeID)
	return promoCodeID, err
}

// UpdatePromoCode обновляет параметры существующего промокода.
// Можно обновить bonusDays, maxUses и/или expiresAt.
func (p *Postgres) UpdatePromoCode(promoCodeID, bonusDays, maxUses int, expiresAt time.Time) error {
	sqlRaw := `
	UPDATE promo_codes
	SET bonus_days = $1, max_uses = $2, expires_at = $3
	WHERE id = $4;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, bonusDays, maxUses, expiresAt, promoCodeID)
	return err
}

// GetAllPromoCodes возвращает список всех промокодов с их статистикой.
func (p *Postgres) GetAllPromoCodes() ([]map[string]any, error) {
	sqlRaw := `
	SELECT id, code, bonus_days, max_uses, used_count, is_active, expires_at, created_at
	FROM promo_codes
	ORDER BY created_at DESC;
	`
	rows, err := p.conn.Query(p.ctx, sqlRaw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any

	for rows.Next() {
		var id int
		var code string
		var bonusDays, maxUses, usedCount int
		var isActive bool
		var expiresAt, createdAt time.Time

		err := rows.Scan(&id, &code, &bonusDays, &maxUses, &usedCount, &isActive, &expiresAt, &createdAt)
		if err != nil {
			return nil, err
		}

		record := map[string]any{
			"id":         id,
			"code":       code,
			"bonus_days": bonusDays,
			"max_uses":   maxUses,
			"used_count": usedCount,
			"is_active":  isActive,
			"expires_at": expiresAt,
			"created_at": createdAt,
		}
		result = append(result, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
