package database

import (
	"database/sql"
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	// Create a temporary SQLite database for testing
	tmpfile, err := os.CreateTemp("", "test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "SQLite Connection",
			config: Config{
				Type:     SQLite,
				FilePath: tmpfile.Name(),
			},
			wantErr: false,
		},
		{
			name: "Invalid Database Type",
			config: Config{
				Type: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Connect(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				defer db.Close()
			}
		})
	}
}

func TestGetTables(t *testing.T) {
	// Create a temporary SQLite database for testing
	tmpfile, err := os.CreateTemp("", "test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Create test database and table
	db, err := sql.Open("sqlite3", tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tables, err := GetTables(db, SQLite)
	if err != nil {
		t.Errorf("GetTables() error = %v", err)
		return
	}

	if len(tables) != 1 || tables[0] != "test_table" {
		t.Errorf("GetTables() = %v, want [test_table]", tables)
	}
}

func TestGetColumns(t *testing.T) {
	// Create a temporary SQLite database for testing
	tmpfile, err := os.CreateTemp("", "test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Create test database and table
	db, err := sql.Open("sqlite3", tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT,
			age INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	columns, err := GetColumns(db, SQLite, "test_table")
	if err != nil {
		t.Errorf("GetColumns() error = %v", err)
		return
	}

	expectedColumns := []string{"id", "name", "age"}
	if len(columns) != len(expectedColumns) {
		t.Errorf("GetColumns() returned %d columns, want %d", len(columns), len(expectedColumns))
		return
	}

	for i, col := range columns {
		if col != expectedColumns[i] {
			t.Errorf("GetColumns()[%d] = %s, want %s", i, col, expectedColumns[i])
		}
	}
}
