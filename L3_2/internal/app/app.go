package app

import (
	"github.com/1lostsun/L3/tree/main/L3_2/internal/config"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/handler"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/repo/postgres"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/usecase"
	"github.com/joho/godotenv"
	"github.com/wb-go/wbf/dbpg"
	"log"
	"os"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DSN")

	opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
	db, err := dbpg.New(dsn, []string{}, opts)
	if err != nil {
		log.Fatal(err)
	}

	config.RunMigrations(dsn)

	pgRepo := postgres.NewRepository(db)
	uc := usecase.NewUseCase(pgRepo)
	h := handler.NewHandler(uc)

	h.InitRoutes()

	if err := h.Run(os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
