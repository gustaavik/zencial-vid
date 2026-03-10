package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/zenfulcode/zencial/internal/infrastructure/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <up|down|status>")
		os.Exit(1)
	}

	command := os.Args[1]

	if err := goose.RunContext(context.Background(), command, db, "migrations"); err != nil {
		log.Fatalf("migration %s failed: %v", command, err)
	}
}
