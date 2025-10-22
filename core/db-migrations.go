package main

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	postgres_migrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func RunDatabaseMigrations() error {

	// Setup db connection
	url := GetConnectionUrl() + "?sslmode=disable"
	db, err := sql.Open("postgres", url)
	if err != nil {
		return fmt.Errorf("Error connecting to the database: %w", err)
	}
	defer db.Close()

	// Check for existing pre-script db
	isExistingDb, err := checkForExistingDatabase(db)
	if err != nil {
		return fmt.Errorf("Error checking if its a pre-script version of the db: %w", err)
	}

	// Setup migrator
	migrator, err := initMigrator(db)
	if err != nil {
		return fmt.Errorf("Error initialising database migrator: %w", err)
	}

	// If it was a pre-script DB, force version to 1
	if isExistingDb {
		log.Info().Msg("Found existing pre-scripted db. Automatically setting as version 1...")
		if err := migrator.Force(1); err != nil {
			return fmt.Errorf("Error forcing migration version to 1 in pre-scripted db: %w", err)
		}
		log.Info().Msg("Successfully marked existing database as migration version 1")
	}

	// Run migrations
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Database migration failed: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Info().Msg("Database schema is up to date, no migrations needed")
	} else {
		version, _, err := migrator.Version()
		if err != nil {
			log.Warn().Msgf("Migration completed, but unable to get current version: %v", err)
		}
		log.Info().Msgf("Successfully migrated database to version: %d", version)
	}

	return nil
}

func initMigrator(db *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres_migrate.WithInstance(db, &postgres_migrate.Config{})
	if err != nil {
		return nil, fmt.Errorf("Error creating migration driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("Error creating migrator instance: %w", err)
	}

	return migrator, nil
}

// Only for users who were using a version of Skystats prior to the db
// creation being scripted. Checks for when the schema_migrations table does not exist,
// but the aircraft_data does. If so, forces the db version to 1.
func checkForExistingDatabase(db *sql.DB) (bool, error) {

	var migrationTableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_migrations'
		)
	`).Scan(&migrationTableExists)

	if err != nil {
		return false, err
	}

	if migrationTableExists {
		return false, nil
	}

	var aircraftTableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'aircraft_data'
		)
	`).Scan(&aircraftTableExists)

	if err != nil {
		return false, err
	}

	return aircraftTableExists, nil
}
