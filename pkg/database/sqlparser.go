package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SQLDumpParser handles parsing of SQL dump files
type SQLDumpParser struct {
	filePath string
	dbType   DBType
	debug    bool
}

// NewSQLDumpParser creates a new SQL dump parser
func NewSQLDumpParser(filePath string, dbType DBType) *SQLDumpParser {
	return &SQLDumpParser{
		filePath: filePath,
		dbType:   dbType,
		debug:    false,
	}
}

// SetDebug enables or disables debug logging
func (p *SQLDumpParser) SetDebug(debug bool) {
	p.debug = debug
}

// logDebug prints a message if debug mode is enabled
func (p *SQLDumpParser) logDebug(format string, args ...interface{}) {
	if p.debug {
		fmt.Printf(format, args...)
	}
}

// ParseToSQLite converts a SQL dump file to a SQLite database
func (p *SQLDumpParser) ParseToSQLite() (string, error) {
	// Create a temporary SQLite database
	tmpfile, err := os.CreateTemp("", "sql_import_*.db")
	if err != nil {
		return "", fmt.Errorf("failed to create temp database: %w", err)
	}
	tmpfile.Close()

	// Connect to the temporary database
	db, err := Connect(Config{
		Type:     SQLite,
		FilePath: tmpfile.Name(),
	})
	if err != nil {
		os.Remove(tmpfile.Name())
		return "", fmt.Errorf("failed to connect to temp database: %w", err)
	}
	defer db.Close()

	// Read and process the SQL dump file
	file, err := os.Open(p.filePath)
	if err != nil {
		os.Remove(tmpfile.Name())
		return "", fmt.Errorf("failed to open SQL dump file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentStatement strings.Builder
	var inCopy bool
	var copyData []string
	var currentTable string
	var inFunction bool
	var inCreateTable bool

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Handle function definitions
		if strings.Contains(line, "CREATE FUNCTION") || strings.Contains(line, "CREATE OR REPLACE FUNCTION") {
			inFunction = true
			continue
		}
		if inFunction {
			if strings.Contains(line, "$$") || strings.Contains(line, "LANGUAGE") {
				inFunction = false
			}
			continue
		}

		// Handle COPY statements
		if strings.HasPrefix(line, "COPY ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				currentTable = strings.TrimPrefix(parts[1], "public.")
				inCopy = true
				copyData = make([]string, 0)
				continue
			}
		}

		if inCopy {
			if line == "\\." {
				// End of COPY data
				inCopy = false
				if err := p.insertCopyData(db, currentTable, copyData); err != nil {
					p.logDebug("Warning: Failed to insert data into %s: %v\n", currentTable, err)
				}
				copyData = nil
				continue
			}
			copyData = append(copyData, line)
			continue
		}

		// Handle CREATE TABLE statements
		if strings.HasPrefix(line, "CREATE TABLE") {
			inCreateTable = true
			line = p.convertCreateTable(line)
		}

		// Handle end of CREATE TABLE
		if inCreateTable && strings.Contains(line, ");") {
			inCreateTable = false
			line = p.cleanupCreateTable(line)
		}

		// Convert syntax for non-CREATE TABLE statements
		if !inCreateTable {
			line = p.convertSyntax(line)
		}

		if line == "" {
			continue
		}

		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")

		if strings.HasSuffix(line, ";") {
			stmt := currentStatement.String()
			if !shouldSkipStatement(stmt) {
				// Execute the statement
				if _, err := db.Exec(stmt); err != nil {
					p.logDebug("Warning: Failed to execute statement: %v\nStatement: %s\n", err, stmt)
				}
			}
			currentStatement.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		os.Remove(tmpfile.Name())
		return "", fmt.Errorf("error reading SQL dump: %w", err)
	}

	return tmpfile.Name(), nil
}

// convertCreateTable handles CREATE TABLE statements specifically
func (p *SQLDumpParser) convertCreateTable(line string) string {
	// Remove schema qualification
	line = strings.ReplaceAll(line, "public.", "")

	// Convert data types
	line = p.convertDataTypes(line)

	return line
}

// cleanupCreateTable cleans up CREATE TABLE statements at their end
func (p *SQLDumpParser) cleanupCreateTable(line string) string {
	// Remove trailing comma before closing parenthesis
	if idx := strings.LastIndex(line, ","); idx != -1 {
		if strings.Index(line[idx:], ")") != -1 {
			line = line[:idx] + line[idx+1:]
		}
	}
	return line
}

// convertDataTypes converts PostgreSQL data types to SQLite types
func (p *SQLDumpParser) convertDataTypes(line string) string {
	conversions := map[string]string{
		"SERIAL":                      "INTEGER",
		"serial":                      "INTEGER",
		"BIGSERIAL":                   "INTEGER",
		"bigserial":                   "INTEGER",
		"SMALLSERIAL":                 "INTEGER",
		"smallserial":                 "INTEGER",
		"timestamp without time zone": "DATETIME",
		"timestamp with time zone":    "DATETIME",
		"boolean":                     "BOOLEAN",
		"BOOLEAN":                     "BOOLEAN",
		"character varying":           "VARCHAR",
		"varchar":                     "VARCHAR",
		"bytea":                       "BLOB",
		"double precision":            "REAL",
		"bigint":                      "INTEGER",
		"int8":                        "INTEGER",
		"smallint":                    "INTEGER",
		"int2":                        "INTEGER",
		"integer":                     "INTEGER",
		"int4":                        "INTEGER",
		"decimal":                     "REAL",
		"numeric":                     "REAL",
		"text":                        "TEXT",
		"json":                        "TEXT",
		"jsonb":                       "TEXT",
		"date":                        "DATE",
	}

	for pg, sqlite := range conversions {
		pattern := fmt.Sprintf(`(?i)%s(\(\d+\))?`, regexp.QuoteMeta(pg))
		re := regexp.MustCompile(pattern)
		line = re.ReplaceAllString(line, sqlite)
	}

	return line
}

// convertSyntax converts database-specific SQL syntax to SQLite syntax
func (p *SQLDumpParser) convertSyntax(line string) string {
	// Skip sequence-related statements
	if strings.Contains(line, "CREATE SEQUENCE") ||
		strings.Contains(line, "ALTER SEQUENCE") ||
		strings.Contains(line, "START WITH") ||
		strings.Contains(line, "SEQUENCE NAME") {
		return ""
	}

	// Skip ownership statements
	if strings.Contains(line, "OWNER TO") {
		return ""
	}

	// Skip PostgreSQL-specific statements
	if strings.Contains(line, "pg_catalog.") {
		return ""
	}

	// Handle constraint statements
	if strings.Contains(line, "ADD CONSTRAINT") {
		return p.convertConstraint(line)
	}

	return line
}

// convertConstraint converts PostgreSQL constraints to SQLite syntax
func (p *SQLDumpParser) convertConstraint(line string) string {
	// Extract constraint details
	if strings.Contains(line, "PRIMARY KEY") {
		matches := regexp.MustCompile(`ADD CONSTRAINT \w+ PRIMARY KEY \((.*?)\)`).FindStringSubmatch(line)
		if len(matches) > 1 {
			tableName := strings.Fields(line)[2] // Get table name from ALTER TABLE statement
			return fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS pk_%s ON %s (%s);",
				tableName, tableName, matches[1])
		}
	}

	if strings.Contains(line, "FOREIGN KEY") {
		matches := regexp.MustCompile(`FOREIGN KEY \((.*?)\) REFERENCES (\w+)\((.*?)\)`).FindStringSubmatch(line)
		if len(matches) > 3 {
			tableName := strings.Fields(line)[2]
			return fmt.Sprintf("CREATE INDEX IF NOT EXISTS fk_%s_%s ON %s (%s);",
				tableName, matches[2], tableName, matches[1])
		}
	}

	if strings.Contains(line, "UNIQUE") {
		matches := regexp.MustCompile(`ADD CONSTRAINT (\w+) UNIQUE \((.*?)\)`).FindStringSubmatch(line)
		if len(matches) > 2 {
			tableName := strings.Fields(line)[2]
			return fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s);",
				matches[1], tableName, matches[2])
		}
	}

	return ""
}

// shouldSkipStatement checks if a statement should be skipped during import
func shouldSkipStatement(stmt string) bool {
	skipPatterns := []string{
		`^SET `,
		`^ALTER DATABASE`,
		`^CREATE DATABASE`,
		`^CREATE SCHEMA`,
		`^CREATE EXTENSION`,
		`^COMMENT ON`,
		`^GRANT `,
		`^REVOKE `,
		`CREATE TRIGGER`,
		`CREATE RULE`,
		`CREATE POLICY`,
		`SELECT pg_catalog`,
		`ALTER TABLE.*OWNER TO`,
		`ALTER TABLE.*ADD GENERATED`,
	}

	for _, pattern := range skipPatterns {
		if matched, _ := regexp.MatchString(pattern, stmt); matched {
			return true
		}
	}

	return false
}

// insertCopyData inserts data from PostgreSQL COPY statements
func (p *SQLDumpParser) insertCopyData(db *sql.DB, table string, data []string) error {
	if len(data) == 0 {
		return nil
	}

	// Begin transaction for faster inserts
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get columns for the table
	columns, err := GetColumns(db, SQLite, table)
	if err != nil {
		return err
	}

	// Prepare the INSERT statement
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert each row
	for _, line := range data {
		values := p.parseCopyLine(line)
		if len(values) != len(columns) {
			continue // Skip invalid rows
		}
		if _, err := stmt.Exec(values...); err != nil {
			p.logDebug("Warning: Failed to insert row into %s: %v\n", table, err)
		}
	}

	return tx.Commit()
}

// parseCopyLine parses a PostgreSQL COPY data line into values
func (p *SQLDumpParser) parseCopyLine(line string) []interface{} {
	var values []interface{}
	fields := strings.Split(line, "\t")

	for _, field := range fields {
		switch field {
		case "\\N":
			values = append(values, nil)
		case "t":
			values = append(values, true)
		case "f":
			values = append(values, false)
		default:
			values = append(values, field)
		}
	}

	return values
}
