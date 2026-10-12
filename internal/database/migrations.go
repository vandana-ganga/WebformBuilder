package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) error {

	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	defer m.Close()

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}