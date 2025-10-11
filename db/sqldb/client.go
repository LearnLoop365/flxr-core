package sqldb

import (
	"context"

	"github.com/LearnLoop365/flxr-core/db"
)

type Client interface {
	db.Client[DBHandle]

	BeginTx(ctx context.Context) (Tx, error)
}
