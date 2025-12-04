package app

import (
	"context"
	"fmt"
	"github.com/1lostsun/L3/internal/config"
	"github.com/1lostsun/L3/internal/repo/pg"
	"github.com/joho/godotenv"
)

func Run() error {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	cfg := config.New()
	db, err := pg.New(ctx, cfg.PgConfig)
	if err != nil {
		return fmt.Errorf("error connecting to postgres: %v", err)
	}

	fmt.Println("Connected to postgres", &db)
	return nil
}
