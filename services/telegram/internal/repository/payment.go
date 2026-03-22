package repository

// NewPayment создает новую запись в таблице payment с chat_id и invoice_payload.
func (p *Postgres) NewPayment(chatID int64, payload string) error {
	sqlRaw := `
	INSERT INTO payment (chat_id, invoice_payload)
	VALUES ($1, $2)
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, chatID, payload)
	return err
}

// SuccessfulPaymentStatus устанавливает successful_payment = true для записи платежа по invoice_payload.
func (p *Postgres) SuccessfulPaymentStatus(payload string) error {
	sqlRaw := `
	UPDATE payment
	SET successful_payment = true
	WHERE invoice_payload = $1
	`
	_, err := p.conn.Exec(p.ctx, sqlRaw, payload)
	return err
}
