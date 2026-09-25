package postgres

import (
	"context"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: ping database: %w", op, err)
	}

	if err := runMigrations(ctx, storagePath); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: run migrations: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	var id int64
	// err := s.db.QueryRow(context.Background(),
	// 	`INSERT INTO url(alias, url) VALUES($1, $2)
	// 	ON CONFLICT (alias) DO UPDATE SET url = EXCLUDED.url
	// 	RETURNING id`,
	// 	alias,
	// 	urlToSave,
	// ).Scan(&id)

	err := s.db.QueryRow(ctx,
		`INSERT INTO url(alias, url) VALUES($1, $2)
		RETURNING id`,
		alias,
		urlToSave,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var url string
	err := s.db.QueryRow(ctx,
		`SELECT url FROM url WHERE alias = $1`, alias).Scan(&url)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) (int64, error) {
	const op = "storage.postgres.DeleteURL"

	result, err := s.db.Exec(ctx,
		"DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement %w", op, err)
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
}

func (s *Storage) UpdateURL(ctx context.Context, alias string, newURL string) (int64, error) {
	const op = "storage.postgres.UpdateURL"

	result, err := s.db.Exec(ctx,
		"UPDATE url SET url = $1 WHERE alias = $2", newURL, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement %w", op, err)
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}
