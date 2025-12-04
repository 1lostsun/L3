package pg

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"log/slog"
)

type Config struct {
	Host     string `envconfig:"PG_USERNAME"`
	Password string `envconfig:"PG_PASSWORD"`
	Username string `envconfig:"PG_HOST"`
	Port     string `envconfig:"PG_PORT"`
	Database string `envconfig:"PG_DATABASE"`
}

func (cfg *Config) PostgresDSN() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	return dsn
}

func RunMigrations(dsn string) {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		slog.Error("failed to create migrator", "error", err)
		return
	}

	if err := m.Up(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		slog.Error(fmt.Sprintf("error running migrations: %v", err))
		return
	}

	slog.Info("Migrations applied successfully")
}
