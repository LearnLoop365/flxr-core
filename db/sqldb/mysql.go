package sqldb

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // side-effect
)

// MysqlClient implements SQLDBClient
type MysqlClient struct {
	Conf *Conf

	// internal fields are implementation details, not exported
	db  *sql.DB
	dsn string
}

func (c *MysqlClient) Init() error {
	var err error
	c.dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=%s&multiStatements=true",
		c.Conf.User,
		c.Conf.PW,
		c.Conf.Host,
		c.Conf.Port,
		c.Conf.DB,
		c.Conf.TZ,
	)
	if c.db, err = sql.Open(c.Conf.Driver, c.dsn); err != nil {
		return err
	}
	c.db.SetConnMaxLifetime(time.Minute * 3)
	c.db.SetMaxOpenConns(10)
	c.db.SetMaxIdleConns(10)
	if err = c.db.Ping(); err != nil {
		return err
	}
	log.Println("[INFO] mysql client initialized")
	return nil
}

func (c *MysqlClient) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

func (c *MysqlClient) DBHandle() *sql.DB {
	return c.db
}
