package database

import (
	"os"
	"strings"
	"testing"
)

func TestSQLDumpParser_ParseToSQLite(t *testing.T) {
	// Create a temporary SQL dump file
	dumpContent := `
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    email VARCHAR(255),
    created_at DATETIME
);

INSERT INTO users (name, email) VALUES
('John Doe', 'john@example.com'),
('Jane Smith', 'jane@example.com');

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    price DECIMAL(10,2)
);
`
	tmpDumpFile, err := os.CreateTemp("", "test_dump_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp dump file: %v", err)
	}
	defer os.Remove(tmpDumpFile.Name())

	if _, err := tmpDumpFile.WriteString(dumpContent); err != nil {
		t.Fatalf("Failed to write dump content: %v", err)
	}
	tmpDumpFile.Close()

	// Create parser and parse the file
	parser := NewSQLDumpParser(tmpDumpFile.Name(), MySQL)
	sqliteDBPath, err := parser.ParseToSQLite()
	if err != nil {
		t.Fatalf("ParseToSQLite() error = %v", err)
	}
	defer os.Remove(sqliteDBPath)

	// Verify the SQLite database was created
	if _, err := os.Stat(sqliteDBPath); os.IsNotExist(err) {
		t.Error("SQLite database file was not created")
	}

	// Connect to the database and verify tables
	db, err := Connect(Config{
		Type:     SQLite,
		FilePath: sqliteDBPath,
	})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite database: %v", err)
	}
	defer db.Close()

	// Check if tables exist
	tables, err := GetTables(db, SQLite)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	expectedTables := []string{"users", "products"}
	if len(tables) != len(expectedTables) {
		t.Errorf("Expected %d tables, got %d", len(expectedTables), len(tables))
	}

	// Verify each expected table exists
	for _, tableName := range expectedTables {
		found := false
		for _, actualTable := range tables {
			if tableName == actualTable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected table %s not found", tableName)
		}
	}
}

func TestConvertSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "MySQL AUTO_INCREMENT",
			input: "id INTEGER PRIMARY KEY AUTO_INCREMENT",
			want:  "id INTEGER PRIMARY KEY AUTOINCREMENT",
		},
		{
			name:  "MySQL ENGINE and CHARSET",
			input: "CREATE TABLE test (id INT) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
			want:  "CREATE TABLE test (id INT);",
		},
		{
			name:  "PostgreSQL SERIAL",
			input: "id SERIAL PRIMARY KEY",
			want:  "id INTEGER PRIMARY KEY AUTOINCREMENT",
		},
		{
			name:  "PostgreSQL timestamp",
			input: "created_at timestamp without time zone",
			want:  "created_at DATETIME",
		},
	}

	parser := NewSQLDumpParser("test.sql", MySQL)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.convertSyntax(tt.input)
			got = strings.TrimSpace(got)
			if got != tt.want {
				t.Errorf("convertSyntax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldSkipStatement(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		want      bool
	}{
		{
			name:      "CREATE TABLE",
			statement: "CREATE TABLE users (id INT);",
			want:      false,
		},
		{
			name:      "USE statement",
			statement: "USE database_name;",
			want:      true,
		},
		{
			name:      "Transaction statement",
			statement: "BEGIN TRANSACTION;",
			want:      true,
		},
		{
			name:      "SET statement",
			statement: "SET foreign_key_checks=0;",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipStatement(tt.statement)
			if got != tt.want {
				t.Errorf("shouldSkipStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLDumpParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		dumpContent string
		wantErr     bool
		errContains string
	}{
		{
			name: "Invalid SQL syntax",
			dumpContent: `
				CREATE TABLE users (
					id INTEGER PRIMARY,
					name TEXT,
				);
			`,
			wantErr:     false, // We don't fail on SQL errors, just log warnings
			errContains: "",
		},
		{
			name: "Valid SQL",
			dumpContent: `
				CREATE TABLE users (
					id INTEGER PRIMARY KEY,
					name TEXT
				);
			`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary dump file
			tmpDumpFile, err := os.CreateTemp("", "test_dump_*.sql")
			if err != nil {
				t.Fatalf("Failed to create temp dump file: %v", err)
			}
			defer os.Remove(tmpDumpFile.Name())

			if _, err := tmpDumpFile.WriteString(tt.dumpContent); err != nil {
				t.Fatalf("Failed to write dump content: %v", err)
			}
			tmpDumpFile.Close()

			// Parse the file
			parser := NewSQLDumpParser(tmpDumpFile.Name(), SQLite)
			sqliteDBPath, err := parser.ParseToSQLite()

			if tt.wantErr {
				if err == nil {
					t.Error("ParseToSQLite() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseToSQLite() error = %v, want error containing %v", err, tt.errContains)
				}
			} else if err != nil {
				t.Errorf("ParseToSQLite() unexpected error = %v", err)
			}

			if sqliteDBPath != "" {
				os.Remove(sqliteDBPath)
			}
		})
	}
}
