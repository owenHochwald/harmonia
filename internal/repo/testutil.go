package repo

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

const (
	testDBHost     = "localhost"
	testDBPort     = "5433" // Test database port
	testDBUser     = "postgres"
	testDBPassword = "secret"
	testDBName     = "harmonia_test"
)

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	mainConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword)

	mainDB, err := sql.Open("postgres", mainConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer mainDB.Close()

	_, err = mainDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	_, err = mainDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	testConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	testDB, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := createTestTables(testDB); err != nil {
		testDB.Close()
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return testDB
}

func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}

func ClearTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec("DELETE FROM fingerprints")
	if err != nil {
		t.Fatalf("Failed to clear fingerprints table: %v", err)
	}

	_, err = db.Exec("DELETE FROM songs")
	if err != nil {
		t.Fatalf("Failed to clear songs table: %v", err)
	}
}

func createTestTables(db *sql.DB) error {
	songsSQL := `
		CREATE TABLE IF NOT EXISTS songs (
			id VARCHAR(255) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			artist VARCHAR(255) NOT NULL,
			album VARCHAR(255),
			year INTEGER,
			s3_key VARCHAR(500) NOT NULL,
			fingerprint BYTEA,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	if _, err := db.Exec(songsSQL); err != nil {
		return fmt.Errorf("failed to create songs table: %w", err)
	}

	fingerprintsSQL := `
		CREATE TABLE IF NOT EXISTS fingerprints (
			id BIGSERIAL PRIMARY KEY,
			song_id VARCHAR(255) NOT NULL,
			hash INTEGER NOT NULL,
			time_offset INTEGER NOT NULL,
			FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
		)
	`

	if _, err := db.Exec(fingerprintsSQL); err != nil {
		return fmt.Errorf("failed to create fingerprints table: %w", err)
	}

	indexSQL := `CREATE INDEX IF NOT EXISTS idx_fingerprints_hash ON fingerprints(hash)`
	if _, err := db.Exec(indexSQL); err != nil {
		return fmt.Errorf("failed to create fingerprints hash index: %w", err)
	}

	return nil
}
