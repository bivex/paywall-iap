package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var databaseURL string
	var migrationsPath string

	flag.StringVar(&databaseURL, "database", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.StringVar(&migrationsPath, "path", "./migrations", "Path to migration files")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Migration setup failed: %v", err)
	}

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Command required: up, down, force")
	}

	command := args[0]
	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully!")
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migrations rolled back successfully!")
	case "migrate":
		if len(args) < 2 {
			log.Fatal("Subcommand required for migrate: up, down")
		}
		if args[1] == "up" {
			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Migration up failed: %v", err)
			}
			fmt.Println("Migrations applied successfully!")
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
