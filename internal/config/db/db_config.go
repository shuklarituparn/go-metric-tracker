package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type Config struct {
	DSN             string        `env:"DATABASE_DSN"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"5m"`
}

func NewDefaultConfig() *Config {
	return &Config{
		DSN:             "",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("err: DSN is required")
	}
	return nil
}

func (c *Config) Connect() (*sql.DB, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("error: invalid DSN %w", err)
	}

	db, err := sql.Open("postgres", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("error: error connecting to the DB: %w", err)
	}
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error: failed to ping database: %w", err)
	}
	return db, nil
}
