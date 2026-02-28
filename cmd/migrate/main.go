package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <command> [args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  up                   Migrate the DB to the most recent version")
		fmt.Println("  up-by-one            Migrate the DB up by 1")
		fmt.Println("  down                 Roll back the version by 1")
		fmt.Println("  redo                 Re-run the latest migration")
		fmt.Println("  reset                Roll back all migrations")
		fmt.Println("  status               Dump the migration status for the current DB")
		fmt.Println("  version              Print the current version of the database")
		fmt.Println("  create NAME [sql|go] Creates new migration file with the current timestamp")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	migrationsDir := "migrations"
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		migrationsDir = dir
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	command := os.Args[1]
	args := os.Args[2:]

	if err := goose.RunContext(ctx, command, db, migrationsDir, args...); err != nil {
		log.Fatalf("Migration %s failed: %v", command, err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}
