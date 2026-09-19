package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Options struct {
	URL      string
	MaxConns int32
	MinConns int32
	Timeout  time.Duration
}

func Open(ctx context.Context, options Options) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(options.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	poolConfig.MaxConns = options.MaxConns
	poolConfig.MinConns = options.MinConns
	poolConfig.ConnConfig.ConnectTimeout = options.Timeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingContext, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	if err := pool.Ping(pingContext); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
