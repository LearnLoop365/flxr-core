package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/LearnLoop365/flxr-core/responses"
)

// QueryRows currently using PREPARE statement
func QueryRows(ctx context.Context, DBHandle *sql.DB, rawStmt string) (*sql.Rows, error) {
	stmt, err := DBHandle.PrepareContext(ctx, rawStmt) // Using PREPARE Statement
	if err != nil {
		return nil, fmt.Errorf("failed to prepare sql statement. %v", err)
	}
	defer func() {
		if stmtCloseErr := stmt.Close(); stmtCloseErr != nil {
			log.Printf("failed to close SQL statement. %v", stmtCloseErr)
		}
	}()
	return stmt.QueryContext(ctx)
}

func RowsToItems[T any](
	rows *sql.Rows,
	fieldPtrsFromItem func(*T) []any, // taking &item, returns []any{&item.Field1, ..., &item.FieldN}
) ([]T, error) {
	var (
		items []T
		err   error
	)
	for rows.Next() {
		var item T
		if err = rows.Scan(fieldPtrsFromItem(&item)...); err != nil {
			return nil, fmt.Errorf("scan failed. %v", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows. %v", err)
	}
	return items, nil
}

func QueryAllItems[T any](
	ctx context.Context,
	DBHandle *sql.DB,
	rawStmt string,
	fieldPtrsFromItem func(*T) []any,
) ([]T, error) {
	rows, err := QueryRows(ctx, DBHandle, rawStmt)
	if err != nil {
		return nil, err
	}
	return RowsToItems[T](rows, fieldPtrsFromItem)
}

func QueryAllItemsResponse[T any](
	DBHandle *sql.DB,
	rawStmt string,
	fieldPtrsFromItem func(*T) []any,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := QueryAllItems[T](r.Context(), DBHandle, rawStmt, fieldPtrsFromItem)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("[ERROR] SQL failed to query items. %v", err))
			return
		}
		responses.EncodeWriteJSON(w, http.StatusOK, items)
	}
}
