package exporter

import (
	"database/sql"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestTableExporter_Export(t *testing.T) {
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

	// Create test table with sample data
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

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO test_table (name, age) VALUES
		('John Doe', 30),
		('Jane Smith', 25),
		('Bob Johnson', 35)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "csv_output")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Create exporter
	columns := []string{"id", "name", "age"}
	exp := NewTableExporter(db, "test_table", columns, outputDir)

	// Export the table
	if err := exp.Export(); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Verify the output file exists
	outputFile := filepath.Join(outputDir, "test_table.csv")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Export() did not create output file")
	}

	// Read and verify the CSV content
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// Verify header
	if len(records) < 1 {
		t.Fatalf("CSV file is empty")
	}
	header := records[0]
	if len(header) != len(columns) {
		t.Errorf("Header length = %d, want %d", len(header), len(columns))
	}
	for i, col := range columns {
		if header[i] != col {
			t.Errorf("Header[%d] = %s, want %s", i, header[i], col)
		}
	}

	// Verify number of records (header + 3 data rows)
	if len(records) != 4 {
		t.Errorf("Number of records = %d, want 4", len(records))
	}

	// Verify data format
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != 3 {
			t.Errorf("Record %d length = %d, want 3", i, len(record))
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "Nil value",
			input: nil,
			want:  "",
		},
		{
			name:  "String value",
			input: "test",
			want:  "test",
		},
		{
			name:  "Integer value",
			input: 42,
			want:  "42",
		},
		{
			name:  "Byte slice",
			input: []byte("test bytes"),
			want:  "test bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.input)
			if got != tt.want {
				t.Errorf("formatValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
