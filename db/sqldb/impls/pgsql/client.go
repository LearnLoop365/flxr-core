package pgsql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LearnLoop365/flxr-core/db/sqldb"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	//sqldb.Client // [Embedded Interface]

	Conf *sqldb.Conf

	// internal fields are implementation details, not exported
	pool *pgxpool.Pool
	dsn  string
}

// Ensure pgsql.Client implements sqldb.Client interface
var _ sqldb.Client = (*Client)(nil)

func (c *Client) Init() error {
	// DSN format for pgx (URL or key/value style)
	// Note: sslmode=disable is often used for local dev, adjust as needed.
	// PostgreSQL natively allows multiple statements in a single query string.
	c.dsn = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&timezone=%s",
		c.Conf.User,
		c.Conf.PW,
		c.Conf.Host,
		c.Conf.Port,
		c.Conf.DB,
		c.Conf.TZ,
	)
	// or key/value DSN format (also supported):
	// c.dsn = fmt.Sprintf(
	//    "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
	//     c.Conf.Host, c.Conf.Port, c.Conf.User, c.Conf.PW, c.Conf.DB, c.Conf.TZ)

	config, err := pgxpool.ParseConfig(c.dsn)
	if err != nil {
		return fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Pool tuning (adjust as needed)
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 3 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect pgx pool: %w", err)
	}

	// connection settings
	if err = c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}

	log.Printf("[INFO] pgsql client initialized (dsn=%s)", c.dsn)
	return nil
}

func (c *Client) DBHandle() sqldb.DBHandle {
	return &DBHandle{pool: c.pool}
}

func (c *Client) Close() error {
	if c.pool == nil {
		return nil
	}
	c.pool.Close()
	log.Println("[INFO] pgsql client closed")
	return nil
}

func (c *Client) BeginTx(ctx context.Context) (sqldb.Tx, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("pgsql client not initialized")
	}
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection failed: %w", err)
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	return &Tx{tx: tx}, nil
}
