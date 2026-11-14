package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("MIGRATION_DATABASE_URL")

	if dsn == "" {
		return errors.New(
			"MIGRATION_DATABASE_URL is required",
		)
	}

	migrationDir := os.Getenv("MIGRATION_DIR")

	if migrationDir == "" {
		migrationDir = "db/migrations"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf(
			"open database: %w", err,
		)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf(
			"database connection failed: %w", err,
		)
	}

	// Never run migrations using a superuser.
	var isSuperuser bool

	err = db.QueryRowContext(
		ctx,
		`SELECT rolsuper
         FROM pg_roles
         WHERE rolname = current_user`,
	).Scan(&isSuperuser)

	if err != nil {
		return fmt.Errorf(
			"check database role: %w", err,
		)
	}

	if isSuperuser {
		return errors.New(
			"refusing to run migrations as a PostgreSQL superuser",
		)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf(
			"configure migration dialect: %w", err,
		)
	}

	fmt.Println("Applying pending database migrations...")

	if err := goose.Up(db, migrationDir); err != nil {
		return fmt.Errorf(
			"apply migrations: %w", err,
		)
	}

	// Migration history is managed by Goose, not by the API.
	// The application does not need access to modify it.
	if _, err := db.Exec(
		`REVOKE ALL PRIVILEGES
         ON TABLE public.goose_db_version
         FROM geomind_app`,
	); err != nil {
		return fmt.Errorf(
			"restrict migration history permissions: %w",
			err,
		)
	}

	fmt.Println("Database bootstrap completed successfully.")

	return nil
}
