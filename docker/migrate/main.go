// tabi-migrate applies ordered, checked-in PostgreSQL migrations. It is kept
// inside the image build boundary so deployment never requires a host tool.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsDirectory = "/app/migrations"

func main() {
	if len(os.Args) != 2 || os.Args[1] != "up" {
		fmt.Fprintln(os.Stderr, "usage: tabi-migrate up")
		os.Exit(2)
	}
	databaseURL, err := secret("TABI_DATABASE_URL")
	if err != nil || databaseURL == "" {
		fmt.Fprintln(os.Stderr, "TABI_DATABASE_URL or TABI_DATABASE_URL_FILE is required")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "database connection configuration is invalid")
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "database is unavailable")
		os.Exit(1)
	}
	if err := migrate(ctx, pool, migrationsDirectory); err != nil {
		fmt.Fprintln(os.Stderr, "migration failed:", err)
		os.Exit(1)
	}
}

func secret(name string) (string, error) {
	if path := strings.TrimSpace(os.Getenv(name + "_FILE")); path != "" {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		value := strings.TrimSpace(string(contents))
		if value == "" {
			return "", errors.New("secret file is empty")
		}
		return value, nil
	}
	return strings.TrimSpace(os.Getenv(name)), nil
}

func migrate(ctx context.Context, pool *pgxpool.Pool, directory string) error {
	// Advisory locks are scoped to a PostgreSQL session. A pool may choose a
	// different session for every call, so keep the checked-out connection for
	// both the lock and every migration operation until it is unlocked.
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock(714812901)`); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlockCtx, `SELECT pg_advisory_unlock(714812901)`)
	}()
	if _, err := connection.Exec(ctx, `CREATE TABLE IF NOT EXISTS public.tabi_schema_migrations (
  filename text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}
	if _, err := connection.Exec(ctx, `DO $$
BEGIN
  IF to_regclass('public.schema_migrations') IS NOT NULL THEN
    INSERT INTO public.tabi_schema_migrations (filename)
    SELECT version FROM public.schema_migrations
    ON CONFLICT (filename) DO NOTHING;
  END IF;
END $$;`); err != nil {
		return fmt.Errorf("reconcile legacy migration ledger: %w", err)
	}
	files, err := filepath.Glob(filepath.Join(directory, "*.up.sql"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("find migrations: %w", err)
	}
	sort.Strings(files)
	for _, file := range files {
		name := filepath.Base(file)
		var applied bool
		if err := connection.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.tabi_schema_migrations WHERE filename = $1)`, name).Scan(&applied); err != nil {
			return fmt.Errorf("check %s: %w", name, err)
		}
		if applied {
			continue
		}
		contents, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		statements := strings.TrimSpace(string(contents))
		// Checked-in migrations include explicit BEGIN/COMMIT wrappers. Remove
		// only that outer pair so the schema change and its ledger row share one
		// transaction; a crash cannot leave an applied but unrecorded migration.
		if strings.HasPrefix(statements, "BEGIN;") && strings.HasSuffix(statements, "COMMIT;") {
			statements = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(statements, "BEGIN;"), "COMMIT;"))
		}
		transaction, err := connection.Begin(ctx)
		if err == nil {
			_, err = transaction.Exec(ctx, statements)
		}
		if err != nil {
			if transaction != nil {
				_ = transaction.Rollback(ctx)
			}
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := transaction.Exec(ctx, `INSERT INTO public.tabi_schema_migrations (filename) VALUES ($1)`, name); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record %s: %w", name, err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}
	return nil
}
