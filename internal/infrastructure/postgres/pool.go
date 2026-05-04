package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(connStr string) (*pgxpool.Pool, error) {
	c, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to parse conn str: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to connect to pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres: failed to ping pool: %w", err)
	}

	return pool, nil
}

func Close(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("postgres: pool is nil")
	}

	done := make(chan struct{})

	go func() {
		pool.Close()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("postgres: pool close timeout: %w", ctx.Err())
	case <-done:
		return nil
	}
}
