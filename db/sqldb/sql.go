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

type RawStmtBank struct {
	stmts map[string]string
}

func (b *RawStmtBank) Get(key string) string {
	return b.stmts[key]
}

func (b *RawStmtBank) All() map[string]string {
	return b.stmts
}

type BankRegEntry struct {
	FS   embed.FS
	Dir  string
	Bank *RawStmtBank
}

var BankRegistry []BankRegEntry

// RegisterBank a model's embedded FS and map
func RegisterBank(bank *RawStmtBank, fs embed.FS, dir string) {
	if bank == nil {
		panic("bank cannot be nil")
	}
	if bank.stmts == nil {
		bank.stmts = make(map[string]string)
	}
	BankRegistry = append(BankRegistry, BankRegEntry{
		FS:   fs,
		Dir:  dir,
		Bank: bank,
	})
}

// LoadRawStmtsForRegisteredBanks registered FS loaders for a given dbtype
// Call this to preload sql files for models after registering the models with sql files
func LoadRawStmtsForRegisteredBanks(dbtype string, placeholderPrefix byte) error {
	modelCnt := 0
	stmtCnt := 0
	for _, loader := range BankRegistry {
		entries, err := loader.FS.ReadDir(loader.Dir)
		if err != nil {
			return fmt.Errorf("failed to read embedded dir: %s. %w", loader.Dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			filename := entry.Name()
			ext := filepath.Ext(filename)
			name := strings.TrimSuffix(filename, ext)
			ext = strings.TrimPrefix(ext, ".")
			data, err := loader.FS.ReadFile(filepath.Join(loader.Dir, filename))
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", filename, err)
			}

			switch ext {
			case dbtype:
				// exact matching file extension -> use it as-is for dialects
				loader.Bank.stmts[name] = string(data)
				stmtCnt++
			case "sql":
				// Standard SQL
				// with Placeholders: `?` (static) and `@` (dynamic)
				if _, exists := loader.Bank.stmts[name]; !exists {
					// Convert static placeholders
					if placeholderPrefix == '?' {
						// Same as our default, skip
						loader.Bank.stmts[name] = string(data)
					} else {
						loader.Bank.stmts[name] = ConvertPlaceholders(string(data), placeholderPrefix)
					}
					stmtCnt++
				}
			}
		}
		modelCnt++
	}
	log.Printf("[INFO] %d sql raw stmts loaded for %d models", stmtCnt, modelCnt)
	return nil
}

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
	rawStmtBank RawStmtBank,
	fieldPtrsFromItem func(*T) []any,
) ([]T, error) {
	rows, err := QueryRows(ctx, DBHandle, rawStmtBank.Get("findall"))
	if err != nil {
		return nil, err
	}
	return RowsToItems[T](rows, fieldPtrsFromItem)
}

func QueryAllItemsResponse[T any](
	DBHandle *sql.DB,
	rawStmtBank RawStmtBank,
	fieldPtrsFromItem func(*T) []any,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := QueryAllItems[T](r.Context(), DBHandle, rawStmtBank, fieldPtrsFromItem)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("[ERROR] SQL failed to query items. %v", err))
			return
		}
		responses.EncodeWriteJSON(w, http.StatusOK, items)
	}
}

type PrepareSource struct {
	Bank *RawStmtBank
	Name string
}
