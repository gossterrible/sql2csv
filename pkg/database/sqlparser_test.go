package database

import (
	"os"
	"strings"
	"testing"
)

func TestSQLDumpParser_ParseToSQLite(t *testing.T) {
	// Create a temporary SQL dump file
	dumpContent := `
-- Test SQL dump
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),
    email VARCHAR(255),
    created_at TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO users (name, email) VALUES
('John Doe', 'john@example.com'),
('Jane Smith', 'jane@example.com');

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),
    price DECIMAL(10,2)
) ENGINE=InnoDB;
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

	for _, expected := range expectedTables {
		found := false
		for _, actual := range tables {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected table %s not found", expected)
		}
	}
}

func TestConvertSyntax(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DBType
		input    string
		expected string
	}{
		{
			name:     "MySQL AUTO_INCREMENT",
			dbType:   MySQL,
			input:    "id INTEGER PRIMARY KEY AUTO_INCREMENT",
			expected: "id INTEGER PRIMARY KEY AUTOINCREMENT",
		},
		{
			name:     "MySQL ENGINE and CHARSET",
			dbType:   MySQL,
			input:    "CREATE TABLE test (id INT) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
			expected: "CREATE TABLE test (id INT)  ;",
		},
		{
			name:     "PostgreSQL SERIAL",
			dbType:   Postgres,
			input:    "id SERIAL PRIMARY KEY",
			expected: "id INTEGER AUTOINCREMENT PRIMARY KEY",
		},
		{
			name:     "PostgreSQL timestamp",
			dbType:   Postgres,
			input:    "created_at timestamp without time zone",
			expected: "created_at DATETIME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &SQLDumpParser{dbType: tt.dbType}
			result := parser.convertSyntax(tt.input)
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("convertSyntax() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldSkipStatement(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{
			name:     "SET statement",
			stmt:     "SET character_set_client = utf8;",
			expected: true,
		},
		{
			name:     "USE statement",
			stmt:     "USE mydatabase;",
			expected: true,
		},
		{
			name:     "CREATE TABLE statement",
			stmt:     "CREATE TABLE test (id INT);",
			expected: false,
		},
		{
			name:     "INSERT statement",
			stmt:     "INSERT INTO test VALUES (1);",
			expected: false,
		},
		{
			name:     "Transaction statement",
			stmt:     "START TRANSACTION;",
			expected: true,
		},
		{
			name:     "CREATE DATABASE statement",
			stmt:     "CREATE DATABASE test;",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipStatement(tt.stmt)
			if result != tt.expected {
				t.Errorf("shouldSkipStatement() = %v, want %v", result, tt.expected)
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
