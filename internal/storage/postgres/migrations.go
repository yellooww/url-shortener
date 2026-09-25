package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func runMigrations(ctx context.Context, databaseURL string) error {
	const op = "storage.postgres.runMigrations"

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("%s: open database: %w", op, err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("%s: ping database: %w", op, err)
	}

	migrationFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("%s: create migrations fs: %w", op, err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrationFS,
	)
	if err != nil {
		return fmt.Errorf("%s: create provider: %w", op, err)
	}
	defer provider.Close()

	_, err = provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("%s: run migrations: %w", op, err)
	}

	return nil
}
