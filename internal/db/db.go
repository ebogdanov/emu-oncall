package db

import (
	"database/sql"
	"fmt"

	"github.com/ebogdanov/emu-oncall/internal/config"
	_ "github.com/lib/pq" // nolint:blank-imports
)

func Init(cfg *config.DB) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Addr, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open(cfg.Driver, connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()

	return db, err
}
