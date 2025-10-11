package sqldb

import "context"

type DBHandle interface {
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	CopyFrom(ctx context.Context, table string, columns []string, rows [][]any) (int64, error)
	Listen(ctx context.Context, channel string) (<-chan Notification, error)

	// InsertStmt - Single INSERT statement, placeholders only
	// to guarantee Result.LastInsertedId() works for auto-increment `id`
	InsertStmt(ctx context.Context, query string, args ...any) (Result, error)
}
