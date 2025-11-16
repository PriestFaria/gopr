package testhelpers

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	abs, err := filepath.Abs("../../migrations")
	if err != nil {
		return fmt.Errorf("abs migrations: %w", err)
	}

	m, err := migrate.New(
		"file://"+abs,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}
