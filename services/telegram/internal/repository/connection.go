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

// ExpiredConnection удаляет все peer с expires_at < NOW() и возвращает список удаленных dto.DelEntity (chat_id, public_key).
func (p *Postgres) ExpiredConnection() ([]dto.DelEntity, error) {
	const sqlRaw = `
	DELETE FROM peer
	WHERE expires_at < NOW()
	RETURNING chat_id, public_key;
	`

	rows, err := p.conn.Query(p.ctx, sqlRaw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.DelEntity

	for rows.Next() {
		var chatID int64
		var publicKey string
		err := rows.Scan(&chatID, &publicKey)
		if err != nil {
			return nil, err
		}
		result = append(result, dto.DelEntity{ChatID: chatID, PublicKey: publicKey})
	}

	// Проверяем, не было ли ошибки при итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

