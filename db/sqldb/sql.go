package sqldb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/LearnLoop365/flxr-core/db"
	"github.com/LearnLoop365/flxr-core/responses"
)

type Conf struct {
	Type   string `json:"type"` // mysql, pgsql, mssql, oracle, maria, sqlite, ...
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Driver string `json:"driver"`
	User   string `json:"user"`
	PW     string `json:"pw"`
	DB     string `json:"db"`
	TZ     string `json:"tz"` // Connection Timezone
}

type Client = db.Client[*sql.DB]

type RawStore struct {
	stmts map[string]string
}

type StoreStmtKey struct {
	Group    string
	StmtName string
}

func (s *RawStore) Set(key StoreStmtKey, rawStmt string) {
	s.stmts[key.Group+"."+key.StmtName] = rawStmt
}

func (s *RawStore) Get(key StoreStmtKey) (string, bool) {
	stmt, exists := s.stmts[key.Group+"."+key.StmtName]
	return stmt, exists
}

func (s *RawStore) GetAll() map[string]string {
	return s.stmts
}

type GroupFS struct {
	Group string
	FS    embed.FS
}

func RegisterGroup(registry *[]GroupFS, fs embed.FS, group string) {
	*registry = append(*registry, GroupFS{
		FS:    fs,
		Group: group,
	})
}

func LoadRawStmtsToStore(registry []GroupFS, store *RawStore, dbtype string, placeholderPrefix byte) error {
	groupCnt := 0
	stmtCnt := 0
	for _, groupFS := range registry {
		files, err := groupFS.FS.ReadDir("sql")
		if err != nil {
			return fmt.Errorf("failed to read embedded `sql` dir. %w", err)
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			filename := f.Name()
			ext := filepath.Ext(filename)
			name := strings.TrimSuffix(filename, ext)
			ext = strings.TrimPrefix(ext, ".")
			data, err := groupFS.FS.ReadFile(filepath.Join("sql", filename))
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", filename, err)
			}
			stmtKey := StoreStmtKey{Group: groupFS.Group, StmtName: name}

			switch ext {
			case dbtype:
				// exact matching file extension -> use it as-is for dialects
				store.Set(stmtKey, string(data))
				stmtCnt++
			case "sql":
				// Standard SQL
				// with Placeholders: `?` (static) and `@` (dynamic)
				if _, exists := store.Get(stmtKey); !exists {
					// Convert static placeholders
					if placeholderPrefix == '?' || placeholderPrefix == 0 {
						// no need to convert
						store.Set(stmtKey, string(data))
					} else {
						store.Set(stmtKey, ConvertStaticPlaceholders(string(data), placeholderPrefix))
					}
					stmtCnt++
				}
			}
		}
		groupCnt++
	}
	log.Printf("[INFO] %d sql raw stmts loaded for %d models", stmtCnt, groupCnt)
	return nil
}

var PlaceholderPrefixForDBType = map[string]byte{
	"mysql":  '?',
	"pgsql":  '$',
	"mssql":  '@',
	"oracle": ':',
	"sqlite": 0, // NOTE: sqlite supports all of them
}

// QueryRows currently using PREPARE statement
// ToDo: make it an option
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
