package pgsql

import (
	"errors"

	"github.com/LearnLoop365/flxr-core/db/sqldb"
	"github.com/jackc/pgx/v5"
)

type Row struct {
	row pgx.Row
}

// Ensure pgsql.Row implements sqldb.Row interface
var _ sqldb.Row = (*Row)(nil)

func (r *Row) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqldb.ErrNoRows
	}
	return err
}
