// Package cache는 Redis 연결을 제공한다.
// 현재 용도: AI 힌트 Rate Limit 카운터 (INCR+Expire 패턴).
// Redis 미연결 시 서버는 정상 동작하고 Rate Limit만 비활성화된다.
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func Connect(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rdb, nil
}
