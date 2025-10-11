package pgsql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/LearnLoop365/flxr-core/db/sqldb"
)

type Client struct {
	//sqldb.Client // [Embedded Interface]

	Conf *sqldb.Conf

	// internal fields are implementation details, not exported
	db  *sql.DB
	dsn string
}

func (c *Client) Init() error {
	var err error

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

	c.db, err = sql.Open("pgx", c.dsn)
	if err != nil {
		return err
	}
	// connection settings
	c.db.SetConnMaxLifetime(3 * time.Minute)
	c.db.SetMaxOpenConns(10)
	c.db.SetMaxIdleConns(10)
	if err = c.db.Ping(); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}
	log.Println("[INFO] pgsql client initialized")
	return nil
}

func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

func (c *Client) DBHandle() *sql.DB {
	return c.db
}
