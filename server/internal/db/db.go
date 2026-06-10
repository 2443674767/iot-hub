package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/user/can-server/config"
)

var DB *sql.DB

func Init(cfg config.DatabaseConfig) error {
	var err error
	DB, err = sql.Open("postgres", cfg.DSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}
