package sqldb

import (
	"context"
	"database/sql"
	"log"
)

func PrepareStmtsInDB(ctx context.Context, dbHandle *sql.DB, prepairedStore map[string]*sql.Stmt, rawStore *RawStore, keys []StoreStmtKey) {
	for _, storeStmtKey := range keys {
		keyStr := storeStmtKey.String()
		rawStmt, ok := rawStore.Get(storeStmtKey)
		if !ok {
			log.Fatalf("[ERROR] raw SQL statement `%s` not found in the rawStore.", keyStr)
		}
		stmt, err := dbHandle.PrepareContext(ctx, rawStmt)
		if err != nil {
			log.Fatalf("[ERROR] failed to prepare statement `%s` in the main DB. %v", keyStr, err)
		}
		prepairedStore[keyStr] = stmt
	}
	log.Printf("[INFO] %d sql stmts prepared", len(keys))
}
