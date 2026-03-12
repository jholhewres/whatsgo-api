package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/postgres/*.sql
var postgresMigrationFS embed.FS

//go:embed migrations/sqlite/*.sql
var sqliteMigrationFS embed.FS

// RunPostgresMigrations executes pending .up.sql migrations for PostgreSQL.
func RunPostgresMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	logger := slog.Default()

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	entries, err := postgresMigrationFS.ReadDir("migrations/postgres")
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, filename := range upFiles {
		version := extractVersion(filename)
		if applied[version] {
			continue
		}

		logger.Info("applying migration", "version", version, "file", filename)

		sqlBytes, err := postgresMigrationFS.ReadFile(filepath.Join("migrations/postgres", filename))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", filename, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("executing migration %s: %w", filename, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("recording migration %s: %w", filename, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("committing migration %s: %w", filename, err)
		}

		logger.Info("applied migration", "version", version)
	}

	return nil
}

// RunSQLiteMigrations executes pending .up.sql migrations for SQLite.
func RunSQLiteMigrations(ctx context.Context, db *sql.DB) error {
	logger := slog.Default()

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	entries, err := sqliteMigrationFS.ReadDir("migrations/sqlite")
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, filename := range upFiles {
		version := extractVersion(filename)
		if applied[version] {
			continue
		}

		logger.Info("applying migration", "version", version, "file", filename)

		sqlBytes, err := sqliteMigrationFS.ReadFile(filepath.Join("migrations/sqlite", filename))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", filename, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning transaction for %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", filename, err)
		}

		logger.Info("applied migration", "version", version)
	}

	return nil
}

func extractVersion(filename string) string {
	name := strings.TrimSuffix(filename, ".up.sql")
	if idx := strings.Index(name, "_"); idx > 0 {
		return name[:idx]
	}
	return name
}
