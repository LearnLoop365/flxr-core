package pgsql

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LearnLoop365/flxr-core/db/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBHandle struct {
	pool *pgxpool.Pool
}

var _ sqldb.DBHandle = (*DBHandle)(nil)

func (h *DBHandle) Query(ctx context.Context, query string, args ...any) (sqldb.Rows, error) {
	rows, err := h.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{
		conn:    nil, // pool manages connection, no need to release here
		current: rows,
		batch:   nil, // single query, no batch
	}, nil
}

func (h *DBHandle) Exec(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	tag, err := h.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Result{tag: tag}, nil
}

func (h *DBHandle) CopyFrom(ctx context.Context, table string, columns []string, rows [][]any) (int64, error) {
	src := pgx.CopyFromRows(rows)
	count, err := h.pool.CopyFrom(ctx, pgx.Identifier{table}, columns, src)
	return count, err
}

func (h *DBHandle) Listen(ctx context.Context, channel string) (<-chan sqldb.Notification, error) {
	conn, err := h.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	notifyCh := make(chan sqldb.Notification)

	go func() {
		defer conn.Release()
		defer close(notifyCh)

		_, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", pgx.Identifier{channel}.Sanitize()))
		if err != nil {
			log.Printf("[WARN] failed to LISTEN on %s: %v", channel, err)
			return
		}

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				log.Printf("[WARN] Listen loop ended for %s: %v", channel, err)
				return
			}
			select {
			case notifyCh <- sqldb.Notification{
				Channel: notification.Channel,
				Payload: notification.Payload,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return notifyCh, nil
}

func (h *DBHandle) InsertStmt(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	trimmed := strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(trimmed), "INSERT") {
		return nil, fmt.Errorf("InsertStmt must start with INSERT")
	}
	// append RETURNING id if missing
	if !strings.Contains(strings.ToUpper(query), "RETURNING") {
		query += " RETURNING id"
		var id int64
		err := h.pool.QueryRow(ctx, query, args...).Scan(&id)
		if err != nil {
			return nil, err
		}
		return &Result{lastInsertID: id}, nil
	}

	tag, err := h.pool.Exec(ctx, query, args...)
	return &Result{tag: tag}, err
}
