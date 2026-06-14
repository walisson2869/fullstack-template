package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"

	"backend/internal/repository/postgres"
)

const migrationsDir = "migrations"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <command> [args]")
		fmt.Fprintln(os.Stderr, "commands: up, up-by-one, down, down-to <version>, reset, status, version, create <name> sql")
		os.Exit(1)
	}

	cfg := postgres.DBConfig{
		Host:     os.Getenv("BLUEPRINT_DB_HOST"),
		Port:     os.Getenv("BLUEPRINT_DB_PORT"),
		Database: os.Getenv("BLUEPRINT_DB_DATABASE"),
		Username: os.Getenv("BLUEPRINT_DB_USERNAME"),
		Password: os.Getenv("BLUEPRINT_DB_PASSWORD"),
		Schema:   os.Getenv("BLUEPRINT_DB_SCHEMA"),
		SSLMode:  os.Getenv("BLUEPRINT_DB_SSLMODE"),
	}

	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("migrate: database: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("migrate: set dialect: %v", err)
	}

	command, args := os.Args[1], os.Args[2:]

	if err := goose.RunContext(context.Background(), command, db, migrationsDir, args...); err != nil {
		log.Fatalf("migrate %s: %v", command, err)
	}
}
