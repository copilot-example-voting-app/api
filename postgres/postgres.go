// Package postgres provides functionality to connect to a postgres database.
package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // https://www.calhoun.io/why-we-import-sql-drivers-with-the-blank-identifier/
)

const (
	// Port is the default port number for postgres.
	Port = 5432

	postgresDriver = "postgres"
)

// Connect creates a new connection to a postgres db.
func Connect(host string, port int, user, passwd, dbname, ssl string) (db *sql.DB, close func() error, err error) {
	sslmode := ""
	if ssl != "" {
		sslmode = fmt.Sprintf("sslmode=%s", ssl)
	}
	source := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s %s",
		host, port, user, passwd, dbname, sslmode)
	db, err = sql.Open(postgresDriver, source)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: sql open: %w", err)
	}
	err = db.Ping() // Force open a connection to the database.
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: ping: %w", err)
	}
	close = func() error {
		err := db.Close()
		if err != nil {
			return fmt.Errorf("postgres: close: %w", err)
		}
		return nil
	}
	return
}
