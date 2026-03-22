package repository

// создает сущность клиента
func (p *Postgres) AddClient(username string, chatID int64) error {
	sqlRaw := `
	INSERT INTO client (username, chat_id, status)
	VALUES ($1, $2, false)
	ON CONFLICT (chat_id) DO NOTHING;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, username, chatID)

	return err
}

// StatusTrue устанавливает status = true для клиента по chat_id.
func (p *Postgres) StatusTrue(chatID int64) error {
	sqlRaw := `
	UPDATE client
	SET status = true
	WHERE chat_id = $1;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID)

	return err
}

// StatusFalse устанавливает status = false для клиента по chat_id.
func (p *Postgres) StatusFalse(chatID int64) error {
	sqlRaw := `
	UPDATE client
	SET status = false
	WHERE chat_id = $1;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID)

	return err
}

// CheckStatus проверяет есть ли у пользователя активная подписка.
// Возвращает true если существует пир с expires_at > NOW().
func (p *Postgres) CheckStatus(chatID int64) (bool, error) {
	sqlRaw := `
	SELECT EXISTS (
		SELECT 1 FROM peer
		WHERE chat_id = $1 AND expires_at > NOW()
	);
	`
	var hasActiveSubscription bool
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&hasActiveSubscription)

	if err != nil {
		return false, err
	}

	return hasActiveSubscription, nil
}

// Tested устанавливает is_tested = true для клиента по chat_id.
func (p *Postgres) Tested(chatID int64) error {
	sqlRaw := `
	UPDATE client
	SET is_tested = true
	WHERE chat_id = $1;
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID)

	return err
}

// IsTested возвращает значение is_tested клиента по chat_id.
func (p *Postgres) IsTested(chatID int64) (bool, error) {
	sqlRaw := `
	SELECT is_tested
	FROM client
	WHERE chat_id = $1
	`
	var isTested bool
	err := p.conn.QueryRow(p.ctx, sqlRaw, chatID).Scan(&isTested)

	if err != nil {
		return false, err
	}

	return isTested, nil
}
