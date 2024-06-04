package db

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	ns    = "go_sql_stats"
	conns = "connections"
)

// StatsGetter is an interface that gets sql.DBStats.
// It's implemented by e.g. *sql.DB or *sqlx.DB.
type StatsGetter interface {
	Stats() sql.DBStats
}

type StatsCollector struct {
	sg StatsGetter

	// descriptions of exported metrics
	maxOpen           *prometheus.Desc
	open              *prometheus.Desc
	inUse             *prometheus.Desc
	idle              *prometheus.Desc
	waitFor           *prometheus.Desc
	blockedSeconds    *prometheus.Desc
	closedMaxIdle     *prometheus.Desc
	closedMaxLifetime *prometheus.Desc
	closedMaxIdleTime *prometheus.Desc
}

func NewStatsCollector(dbName string, sg StatsGetter) *StatsCollector {
	labels := prometheus.Labels{"db_name": dbName}
	return &StatsCollector{
		sg: sg,
		maxOpen: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "max_open"),
			"Maximum number of open connections to the database.",
			nil,
			labels,
		),
		open: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "open"),
			"The number of established connections both in use and idle.",
			nil,
			labels,
		),
		inUse: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "in_use"),
			"The number of connections currently in use.",
			nil,
			labels,
		),
		idle: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "idle"),
			"The number of idle connections.",
			nil,
			labels,
		),
		waitFor: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "waited_for"),
			"The total number of connections waited for.",
			nil,
			labels,
		),
		blockedSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "blocked_seconds"),
			"The total time blocked waiting for a new connection.",
			nil,
			labels,
		),
		closedMaxIdle: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "closed_max_idle"),
			"The total number of connections closed due to SetMaxIdleConns.",
			nil,
			labels,
		),
		closedMaxLifetime: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "closed_max_lifetime"),
			"The total number of connections closed due to SetConnMaxLifetime.",
			nil,
			labels,
		),
		closedMaxIdleTime: prometheus.NewDesc(
			prometheus.BuildFQName(ns, conns, "closed_max_idle_time"),
			"The total number of connections closed due to SetConnMaxIdleTime.",
			nil,
			labels,
		),
	}
}

func (c StatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpen
	ch <- c.open
	ch <- c.inUse
	ch <- c.idle
	ch <- c.waitFor
	ch <- c.blockedSeconds
	ch <- c.closedMaxIdle
	ch <- c.closedMaxLifetime
	ch <- c.closedMaxIdleTime
}

func (c StatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.sg.Stats()

	ch <- prometheus.MustNewConstMetric(
		c.maxOpen,
		prometheus.GaugeValue,
		float64(stats.MaxOpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		c.open,
		prometheus.GaugeValue,
		float64(stats.OpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		c.inUse,
		prometheus.GaugeValue,
		float64(stats.InUse),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idle,
		prometheus.GaugeValue,
		float64(stats.Idle),
	)
	ch <- prometheus.MustNewConstMetric(
		c.waitFor,
		prometheus.CounterValue,
		float64(stats.WaitCount),
	)
	ch <- prometheus.MustNewConstMetric(
		c.blockedSeconds,
		prometheus.CounterValue,
		stats.WaitDuration.Seconds(),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxIdle,
		prometheus.CounterValue,
		float64(stats.MaxIdleClosed),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxLifetime,
		prometheus.CounterValue,
		float64(stats.MaxLifetimeClosed),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxIdleTime,
		prometheus.CounterValue,
		float64(stats.MaxIdleTimeClosed),
	)
}
