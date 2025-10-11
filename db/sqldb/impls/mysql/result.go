package mysql

import (
	"database/sql"

	"github.com/LearnLoop365/flxr-core/db/sqldb"
)

type Result struct {
	result sql.Result
}

// Ensure mysql.Result implements sqldb.Result interface
var _ sqldb.Result = (*Result)(nil)

func (r *Result) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

func (r *Result) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}
