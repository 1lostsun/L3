package config

import (
	"github.com/1lostsun/L3/internal/repo/pg"
	"os"
)

type Config struct {
	PgConfig *pg.Config
}

func New() *Config {
	var cfg pg.Config
	cfg.Host = os.Getenv("PG_HOST")
	cfg.Port = os.Getenv("PG_PORT")
	cfg.Username = os.Getenv("PG_USERNAME")
	cfg.Password = os.Getenv("PG_PASSWORD")
	cfg.Database = os.Getenv("PG_DATABASE")

	return &Config{PgConfig: &cfg}
}
