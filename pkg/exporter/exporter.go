package exporter

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const batchSize = 1000

// TableExporter handles the export of a single table to CSV
type TableExporter struct {
	db        *sql.DB
	tableName string
	columns   []string
	output    string
}

// NewTableExporter creates a new TableExporter instance
func NewTableExporter(db *sql.DB, tableName string, columns []string, outputDir string) *TableExporter {
	return &TableExporter{
		db:        db,
		tableName: tableName,
		columns:   columns,
		output:    filepath.Join(outputDir, fmt.Sprintf("%s.csv", tableName)),
	}
}

// Export exports the table to a CSV file
func (e *TableExporter) Export() error {
	file, err := os.Create(e.output)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(e.columns); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Prepare the query
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(e.columns, ", "),
		e.tableName)

	rows, err := e.db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying data: %w", err)
	}
	defer rows.Close()

	// Prepare the value holders for scanning
	values := make([]interface{}, len(e.columns))
	valuePtrs := make([]interface{}, len(e.columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Process rows in batches
	batch := make([][]string, 0, batchSize)
	count := 0

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		// Convert values to strings
		record := make([]string, len(e.columns))
		for i, val := range values {
			record[i] = formatValue(val)
		}

		batch = append(batch, record)
		count++

		if count >= batchSize {
			if err := writer.WriteAll(batch); err != nil {
				return fmt.Errorf("error writing batch: %w", err)
			}
			batch = batch[:0]
			count = 0
		}
	}

	// Write remaining records
	if len(batch) > 0 {
		if err := writer.WriteAll(batch); err != nil {
			return fmt.Errorf("error writing final batch: %w", err)
		}
	}

	return nil
}

// formatValue converts an interface{} to a string representation
func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch v := v.(type) {
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
} 