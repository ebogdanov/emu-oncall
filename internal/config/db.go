package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultDBAddr         = "127.0.0.1"
	defaultDBPort         = 5432
	defaultDBUser         = "admin"
	defaultDBPassword     = "admin"
	defaultDBName         = "emu_oncall"
	defaultDBDriver       = "postgres"
	defaultDBTimeout      = 300 * time.Millisecond
	defaultDBReadTimeout  = 50 * time.Millisecond
	defaultDBWriteTimeout = 100 * time.Millisecond
)

// Default connection pool configuration
const (
	defaultDBMaxOpenConns    = 100
	defaultDBMaxIdleConns    = 100
	defaultDBConnMaxLifetime = 0
)

func d(opt string) string {
	return fmt.Sprintf("db.%s", opt)
}

type DB struct {
	Addr     string
	Port     int
	User     string
	Password string
	DBName   string
	Driver   string

	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration

	Retry Retry
}

type Retry struct {
	RetryMax     int
	RetryTimeout time.Duration
}

func ParseDB() *DB {
	viper.SetDefault(d("addr"), defaultDBAddr)
	viper.SetDefault(d("port"), defaultDBPort)
	viper.SetDefault(d("user"), defaultDBUser)
	viper.SetDefault(d("password"), defaultDBPassword)
	viper.SetDefault(d("dbname"), defaultDBName)
	viper.SetDefault(d("driver"), defaultDBDriver)

	viper.SetDefault(d("timeout"), defaultDBTimeout)
	viper.SetDefault(d("read_timeout"), defaultDBReadTimeout)
	viper.SetDefault(d("write_timeout"), defaultDBWriteTimeout)

	viper.SetDefault(d("pool.max_open_connections"), defaultDBMaxOpenConns)
	viper.SetDefault(d("pool.max_idle_connections"), defaultDBMaxIdleConns)
	viper.SetDefault(d("pool.max_life_time"), defaultDBConnMaxLifetime)

	cfg := &DB{
		Addr: viper.GetString(d("addr")),
		Port: viper.GetInt(d("port")),

		User:     viper.GetString(d("user")),
		Password: viper.GetString(d("password")),
		DBName:   viper.GetString(d("dbname")),
		Driver:   viper.GetString(d("driver")),

		Timeout:      viper.GetDuration(d("timeout")),
		ReadTimeout:  viper.GetDuration(d("read_timeout")),
		WriteTimeout: viper.GetDuration(d("write_timeout")),

		MaxOpenConnections:    viper.GetInt(d("pool.max_open_connections")),
		MaxIdleConnections:    viper.GetInt(d("pool.max_idle_connections")),
		ConnectionMaxLifetime: viper.GetDuration(d("pool.max_life_time")),
	}

	return cfg
}
