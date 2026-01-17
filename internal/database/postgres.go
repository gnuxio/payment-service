package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")

	return &DB{db}, nil
}

func (db *DB) RunMigrations() error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Read all migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	// Sort migrations by name
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Run each migration
	for _, filename := range migrationFiles {
		// Check if migration has already been applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking migration status: %w", err)
		}

		if exists {
			log.Printf("Migration %s already applied, skipping", filename)
			continue
		}

		// Read migration file
		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", filename, err)
		}

		// Execute migration
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("error executing migration %s: %w", filename, err)
		}

		// Record migration as applied
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", filename)
		if err != nil {
			return fmt.Errorf("error recording migration %s: %w", filename, err)
		}

		log.Printf("Successfully applied migration: %s", filename)
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
