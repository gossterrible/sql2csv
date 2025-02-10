package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sql2csv/pkg/cli"
	"sql2csv/pkg/database"
	"sql2csv/pkg/exporter"
	"strings"
	"sync"
)

func main() {
	// Get database configuration from user
	config, err := cli.DatabaseConfig()
	if err != nil {
		log.Fatalf("Error getting database configuration: %v", err)
	}

	// Clean up temporary SQLite database if using SQL dump
	if config.Type == database.SQLite && strings.Contains(config.FilePath, "sql_import_") {
		defer os.Remove(config.FilePath)
	}

	// Connect to the database
	db, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Let user select tables to export
	selectedTables, err := cli.SelectTables(db, config.Type)
	if err != nil {
		log.Fatalf("Error selecting tables: %v", err)
	}

	// Get output directory
	outputDir, err := cli.SelectOutputDir()
	if err != nil {
		log.Fatalf("Error selecting output directory: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Create a wait group to handle concurrent exports
	var wg sync.WaitGroup
	// Create an error channel to collect errors from goroutines
	errChan := make(chan error, len(selectedTables))

	// Export each selected table
	for _, table := range selectedTables {
		wg.Add(1)
		go func(tableName string) {
			defer wg.Done()

			// Get columns for the table
			columns, err := database.GetColumns(db, config.Type, tableName)
			if err != nil {
				errChan <- fmt.Errorf("error getting columns for table %s: %v", tableName, err)
				return
			}

			// Create exporter for the table
			exp := exporter.NewTableExporter(db, tableName, columns, outputDir)

			// Export the table
			if err := exp.Export(); err != nil {
				errChan <- fmt.Errorf("error exporting table %s: %v", tableName, err)
				return
			}

			fmt.Printf("Successfully exported table %s to %s\n",
				tableName, filepath.Join(outputDir, tableName+".csv"))
		}(table)
	}

	// Wait for all exports to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	hasErrors := false
	for err := range errChan {
		if err != nil {
			hasErrors = true
			log.Printf("Error during export: %v\n", err)
		}
	}

	if !hasErrors {
		fmt.Println("\nAll tables exported successfully!")
	}
}
