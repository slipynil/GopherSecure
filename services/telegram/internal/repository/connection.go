package repository

import (
	"telegram-service/internal/dto"
	"time"
)

// SaveConnection сохраняет полную информацию о подключении пира в одной операции.
// Удаляет старые записи для этого chat_id и создает новую с обоими ключами и сроком истечения.
func (p *Postgres) SaveConnection(chatID int64, publicKey, presharedKey string, expiresAt time.Time) error {
	sqlRaw := `
	DELETE FROM peer WHERE chat_id = $1;

	INSERT INTO peer (chat_id, public_key, preshared_key, expires_at)
	VALUES ($1, $2, $3, $4);
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID, publicKey, presharedKey, expiresAt)
	return err
}

// GetHostID возвращает host_id peer'а по chat_id.
func (p *Postgres) GetHostID(chatID int64) (int, error) {
	sqlRaw := `
	SELECT host_id FROM peer
	WHERE chat_id = $1;
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

