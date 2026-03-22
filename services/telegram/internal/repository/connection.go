package repository

import (
	"telegram-service/internal/dto"
	"time"
)

// SaveConnection обновляет существующую запись пира по host_id с обоими ключами и сроком истечения.
// Ожидается, что запись уже создана вызовом NewConnection(), который вернул host_id.
func (p *Postgres) SaveConnection(hostID int, publicKey, presharedKey string, expiresAt time.Time) error {
	sqlRaw := `
	UPDATE peer
	SET public_key = $2, preshared_key = $3, expires_at = $4
	WHERE host_id = $1;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, hostID, publicKey, presharedKey, expiresAt)
	return err
}

// NewConnection создает новую запись в таблице peer с chat_id и возвращает автогенерированный host_id.
// Устанавливает временный expires_at, который будет обновлен в SaveConnection().
func (p *Postgres) NewConnection(chatID int64) (int, error) {
	sqlRaw := `
	INSERT INTO peer (chat_id, expires_at)
	VALUES ($1, NOW() + INTERVAL '24 hours')
	RETURNING host_id;
	`
	var hostID int
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&hostID)
	return hostID, err
}

// GetHostID возвращает host_id peer'а по chat_id.
func (p *Postgres) GetHostID(chatID int64) (int, error) {
	sqlRaw := `
	SELECT host_id FROM peer
	WHERE chat_id = $1
	ORDER BY host_id DESC
	LIMIT 1;
	`
	var hostID int
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&hostID)
	return hostID, err
}

// DeleteConnection удаляет запись пира из таблицы peer по host_id.
func (p *Postgres) DeleteConnection(hostID int) error {
	sqlRaw := `DELETE FROM peer WHERE host_id = $1;`
	_, err := p.conn.Exec(p.ctx, sqlRaw, hostID)
	return err
}

// GetConnection возвращает ключи и virtual socket существующего пира по chat_id.
// Используется при продлении подписки для восстановления пира в WireGuard.
func (p *Postgres) GetConnection(chatID int64) (*dto.DelEntity, error) {
	sqlRaw := `
	SELECT host_id, chat_id, public_key, preshared_key
	FROM peer
	WHERE chat_id = $1
	ORDER BY host_id DESC
	LIMIT 1;
	`
	var e dto.DelEntity
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&e.HostID, &e.ChatID, &e.PublicKey, &e.PresharedKey)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// MarkExpired помечает запись как обработанную — устанавливает expires_at в эпоху.
// Пир деактивирован в WireGuard, но ключи сохраняются для восстановления при продлении.
func (p *Postgres) MarkExpired(hostID int) error {
	sqlRaw := `UPDATE peer SET expires_at = 'epoch' WHERE host_id = $1;`
	_, err := p.conn.Exec(p.ctx, sqlRaw, hostID)
	return err
}

// ExpiredConnection возвращает список пиров с истекшей активной подпиской.
// Исключает уже обработанные записи (expires_at = 'epoch').
func (p *Postgres) ExpiredConnection() ([]dto.DelEntity, error) {
	const sqlRaw = `
	SELECT chat_id, public_key, preshared_key, host_id
	FROM peer
	WHERE expires_at < NOW() AND expires_at > 'epoch';
	`

	rows, err := p.conn.Query(p.ctx, sqlRaw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.DelEntity

	for rows.Next() {
		var e dto.DelEntity
		err := rows.Scan(&e.ChatID, &e.PublicKey, &e.PresharedKey, &e.HostID)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}

	// Проверяем, не было ли ошибки при итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// HasPeerWithKeys проверяет, есть ли у пользователя пир с заполненными ключами.
// Не проверяет срок действия подписки, просто наличие публичного и preshared ключей.
func (p *Postgres) HasPeerWithKeys(chatID int64) (bool, error) {
	sqlRaw := `
	SELECT EXISTS (
		SELECT 1 FROM peer
		WHERE chat_id = $1
		AND public_key IS NOT NULL
		AND public_key != ''
		AND preshared_key IS NOT NULL
		AND preshared_key != ''
	);
	`
	var exists bool
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&exists)
	return exists, err
}

