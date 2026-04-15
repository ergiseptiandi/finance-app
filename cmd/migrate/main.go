package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"finance-backend/internal/config"
	"finance-backend/internal/database"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("no .env file loaded; using system environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.ValidateForMigrate(); err != nil {
		log.Fatal(err)
	}

	command, args, err := parseCommand(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.OpenMySQLForMigration(ctx, cfg.MySQL)
	if err != nil {
		log.Fatalf("mysql connection failed: %v", err)
	}
	defer db.Close()

	driver, err := mysqlmigrate.WithInstance(db, &mysqlmigrate.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		log.Fatal(err)
	}

	sourceURL, err := migrationsSourceURL("migrations")
	if err != nil {
		log.Fatal(err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, cfg.MySQL.Name, driver)
	if err != nil {
		log.Fatal(err)
	}
	defer migrator.Close()

	if err := runCommand(migrator, command, args); err != nil {
		log.Fatal(err)
	}
}

func parseCommand(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("usage: go run ./cmd/migrate [up|down|version|force]")
	}

	return args[0], args[1:], nil
}

func runCommand(migrator *migrate.Migrate, command string, args []string) error {
	switch command {
	case "up":
		return ignoreNoChange(migrator.Up())
	case "down":
		steps := 1
		if len(args) > 0 {
			parsed, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid down steps: %w", err)
			}
			steps = parsed
		}

		return ignoreNoChange(migrator.Steps(-steps))
	case "version":
		version, dirty, err := migrator.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Print("migration version: none")
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("migration version: %d dirty=%t", version, dirty)
		return nil
	case "force":
		if len(args) == 0 {
			return fmt.Errorf("usage: go run ./cmd/migrate force <version>")
		}

		version, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid force version: %w", err)
		}

		return migrator.Force(version)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func ignoreNoChange(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		log.Print("no change")
		return nil
	}

	return err
}

func migrationsSourceURL(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	slashed := filepath.ToSlash(absDir)
	if !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}

	return (&url.URL{Scheme: "file", Path: slashed}).String(), nil
}
