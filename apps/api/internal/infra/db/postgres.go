// Package db는 PostgreSQL 커넥션 풀과 마이그레이션을 제공한다.
// pgxpool을 사용해 고루틴 안전한 커넥션 풀 관리.
package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pg ping: %w", err)
	}
	return pool, nil
}

// MigrateUp runs the init migration from the migrations directory.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool) error {
	sql, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}
