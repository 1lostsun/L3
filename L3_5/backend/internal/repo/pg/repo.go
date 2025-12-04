package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type Repo struct {
	pg *pgxpool.Pool
}

func New(ctx context.Context, cfg *Config) (*Repo, error) {
	dsn := cfg.PostgresDSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("failed to parse pg config")
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connected successfully")
	return &Repo{pg: pool}, nil
}
