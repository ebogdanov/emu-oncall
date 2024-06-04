package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	heartBeatCount          = "emu_heartbeat_request_count"
	externalAPIDurationTime = "emu_external_api_request_duration"
	notificationsCount      = "emu_notifications_request_count"
	httpRequestDuration     = "http_request_duration"
	scheduleCount           = "emu_read_schedule_count"
	dbQueryDuration         = "db_query_duration"
	dbQueryErrorsCount      = "db_query_errors_count"
)

type Storage struct {
	Heartbeat           prometheus.Gauge
	ResponseTime        prometheus.Histogram
	APIResponseTime     *prometheus.HistogramVec
	DBQueryTime         *prometheus.HistogramVec
	Notifications       *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	SchedulesCounter    *prometheus.CounterVec
	DBQueryErrorCounter *prometheus.CounterVec
}

func NewMetrics() *Storage {
	m := &Storage{
		Heartbeat: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: heartBeatCount,
			Help: "Count of calls for Heartbeat endpoint",
		}),
		Notifications: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: notificationsCount,
				Help: "Count of notifications",
			},
			[]string{"type"},
		),
		APIResponseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    externalAPIDurationTime,
				Help:    "Duration of API requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint", "method", "http_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    httpRequestDuration,
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"handler", "method", "http_code"},
		),
		SchedulesCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: scheduleCount,
				Help: "Update schedules stats",
			},
			[]string{"name", "status"},
		),
		DBQueryTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    dbQueryDuration,
				Help:    "Duration of DB queries in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query"},
		),
		DBQueryErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: dbQueryErrorsCount,
				Help: "DB queries errors",
			},
			[]string{"query"},
		),
	}

	prometheus.MustRegister(m.Heartbeat)
	prometheus.MustRegister(m.Notifications)
	prometheus.MustRegister(m.APIResponseTime)
	prometheus.MustRegister(m.HTTPRequestDuration)
	prometheus.MustRegister(m.SchedulesCounter)
	prometheus.MustRegister(m.DBQueryErrorCounter)
	prometheus.MustRegister(m.DBQueryTime)

	return m
}
