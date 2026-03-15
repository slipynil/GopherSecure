package repository

import (
	"errors"
	"telegram-service/internal/dto"
	"time"

	"github.com/jackc/pgx/v5"
)

func (p *Postgres) NewConnection(chatID int64, expiresAt time.Time) (int, error) {
	sqlRaw := `INSERT INTO peer (chat_id, expires_at) VALUES ($1, $2) RETURNING host_id`
	var hostID int
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID, expiresAt).Scan(&hostID)
	return hostID, err
}

func (p *Postgres) DeleteConnection(chatID int64) error {
	sqlRaw := `DELETE FROM peer WHERE chat_id = $1`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID)
	return err
}

func (p *Postgres) RenewConnection(chatID int64, expiresAt time.Time) error {
	sqlRaw := `UPDATE peer SET expires_at = $2 WHERE chat_id = $1`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID, expiresAt)
	return err
}

func (p *Postgres) GetPeer(chatID int64) (publicKey, presharedKey string, err error) {
	sqlRaw := `SELECT COALESCE(public_key, ''), COALESCE(preshared_key, '') FROM peer WHERE chat_id = $1`
	err = p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&publicKey, &presharedKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", dto.ErrNotFound
	}
	return
}

func (p *Postgres) SaveKeys(chatID int64, pubKey, psk string) error {
	sqlRaw := `UPDATE peer SET public_key = $2, preshared_key = $3 WHERE chat_id = $1`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID, pubKey, psk)
	return err
}

func (p *Postgres) ExpiredConnection() ([]dto.DelEntity, error) {
	const sqlRaw = `
	DELETE FROM peer
	WHERE expires_at < NOW()
	RETURNING chat_id, public_key, preshared_key;
	`

	rows, err := p.conn.Query(p.ctx, sqlRaw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.DelEntity
	for rows.Next() {
		var entity dto.DelEntity
		if err := rows.Scan(&entity.ChatID, &entity.PublicKey, &entity.PresharedKey); err != nil {
			return nil, err
		}
		result = append(result, entity)
	}
	return result, rows.Err()
}
