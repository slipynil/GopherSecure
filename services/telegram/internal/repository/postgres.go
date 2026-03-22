package repository

// Package repository предоставляет реализацию репозитория для PostgreSQL с методами работы с таблицами client, payment и peer в Telegram-боте.

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Postgres представляет подключение к базе PostgreSQL.
type Postgres struct {
	conn *pgx.Conn
	ctx  context.Context
}

// New открывает соединение с PostgreSQL по указанному DSN и возвращает объект Postgres.
func New(ctx context.Context, dbConn string) (*Postgres, error) {
	conn, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		conn: conn,
		ctx:  ctx,
	}, nil
}

// Close закрывает соединение с базой данных.
func (p *Postgres) Close() error {
	return p.conn.Close(p.ctx)
}
