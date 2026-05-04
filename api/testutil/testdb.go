// Package testutil provides shared test helpers for integration tests.
package testutil

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// NewTestDB spins up a PostgreSQL Docker container, runs schema migrations,
// and returns the database handle and a cleanup function.
//
// Usage:
//
//	func TestMain(m *testing.M) {
//	    db, cleanup := testutil.NewTestDB(m)
//	    defer cleanup()
//	    testDB = db
//	    os.Exit(m.Run())
//	}
func NewTestDB(m *testing.M) (*sql.DB, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseURL)

	resource.Expire(120)

	var db *sql.DB
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Run schema migrations
	schemaPath := "./sqlc/schema.sql"
	// Also try parent path for sub-package tests
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "../sqlc/schema.sql"
	}

	file, err := os.Open(schemaPath)
	if err != nil {
		log.Fatalf("Could not open schema file: %s", err)
	}
	defer file.Close()

	schemaBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Could not read schema file: %s", err)
	}

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		log.Fatalf("Could not execute SQL: %s", err)
	}

	cleanup := func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}

	return db, cleanup
}
