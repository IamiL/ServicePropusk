package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"database_name"`
	User     string `yaml:"username"`
	Pass     string `yaml:"password"`
	MaxConns int    `yaml:"max_connections" default:"10"`
}

func (c *Config) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.DBName,
	)
}

func NewConnPool(config *Config) (*pgxpool.Pool, error) {
	pgxPollConfig, err := pgxpool.ParseConfig(config.ConnString())
	if err != nil {
		log.Fatal("Ошибка конфигурации пула: ", err)
	}

	pgxPollConfig.MaxConns = int32(config.MaxConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxPollConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Ошибка создания пула: %w", err))
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Ошибка ping к postgres: %w", err))
	}

	return pool, err
}
