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
	poolConfig, err := pgxpool.ParseConfig(options.URL) // 解析DSN,拿到一份配置
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	poolConfig.MaxConns = options.MaxConns                 // 池上限
	poolConfig.MinConns = options.MinConns                 // 池下=下限(保活)
	poolConfig.ConnConfig.ConnectTimeout = options.Timeout // 单条连接的建立超时

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig) // 建池(此时还没有建立连接)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingContext, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	if err := pool.Ping(pingContext); err != nil { //真正拨一次,验证可达
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
