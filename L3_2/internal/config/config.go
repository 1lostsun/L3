package config

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func RunMigrations(dsn string) {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		log.Fatal(err)
	}
	log.Println("Migrations applied successfully")
}
