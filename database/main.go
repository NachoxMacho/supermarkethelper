package database

import (
	"database/sql"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	// We don't use any of the functions in these, they are just providers
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)


func Initialize(driverName string, path string) error {

	slog.Info("Initializing database", slog.String("driver", driverName), slog.String("path", path))

	db, err := sql.Open(driverName, path)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	persistentDB = db

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver)
	if err != nil {
		return err
	}
	log.Println("Running Migrations")
	err = m.Up()
	// Need to handle no change err
	// https://github.com/golang-migrate/migrate/issues/485
	if err != nil && err != migrate.ErrNoChange {
		log.Println("Migration Error")
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	log.Println("Database is on version:", version)

	// This isn't technically an error, but isn't great
	if dirty {
		log.Println("Database is dirty, FIX IT NOW")
	}

	return nil
}

var persistentDB *sql.DB

func ConnectDB() (*sql.DB) {
	return persistentDB
}

