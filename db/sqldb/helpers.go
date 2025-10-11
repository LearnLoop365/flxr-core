package sqldb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LearnLoop365/flxr-core/responses"
)

func RowsToItems[T any](
	rows Rows,
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
	DBHandle DBHandle,
	rawStmt string,
	fieldPtrsFromItem func(*T) []any,
) ([]T, error) {
	rows, err := DBHandle.QueryRows(ctx, rawStmt)
	if err != nil {
		return nil, err
	}
	return RowsToItems[T](rows, fieldPtrsFromItem)
}

func QueryAllItemsResponse[T any](
	DBHandle DBHandle,
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
