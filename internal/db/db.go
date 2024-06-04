package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ebogdanov/emu-oncall/internal/config"
	_ "github.com/lib/pq" // nolint:blank-imports
)

type DBx struct {
	pm *metrics.Storage
	*sql.DB
}

var letterToQuery = map[string]string{
	"u": "UPDATE",
	"s": "SELECT",
	"d": "DELETE",
	"i": "INSERT",
}

func (db *DBx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()

	res, err := db.DB.QueryContext(ctx, query, args...)

	duration := time.Since(start)

	// Dumb way to categorize query by type
	queryType := "OTHER"
	if len(query) >= 1 {
		letter := strings.ToLower(query[:1])

		if _, ok := letterToQuery[letter]; ok {
			queryType = letterToQuery[letter]
		}
	}

	db.pm.DBQueryTime.WithLabelValues(queryType).Observe(duration.Seconds())

	if err != nil {
		db.pm.DBQueryErrorCounter.WithLabelValues(queryType).Inc()
	}

	return res, err
}

func Init(cfg *config.DB, pm *metrics.Storage) (*DBx, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Addr, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	dbX := &DBx{
		pm: pm,
	}

	db, err := sql.Open(cfg.Driver, connStr)
	if err != nil {
		return nil, err
	}

	collector := NewStatsCollector(cfg.DBName, db)
	prometheus.MustRegister(collector)

	dbX.DB = db

	return dbX, db.Ping()
}
