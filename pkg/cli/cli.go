package cli

import (
	"database/sql"
	"fmt"
	"sql2csv/pkg/database"

	"github.com/AlecAivazis/survey/v2"
)

// DatabaseConfig prompts the user for database connection details
func DatabaseConfig() (database.Config, error) {
	var config database.Config
	var dbTypeStr string

	// First, ask if using SQL dump file or direct connection
	var connectionType string
	connectionPrompt := &survey.Select{
		Message: "Select connection type:",
		Options: []string{"Direct Connection", "Connection String", "SQL Dump File"},
	}
	if err := survey.AskOne(connectionPrompt, &connectionType); err != nil {
		return config, err
	}

	// Handle SQL dump file
	if connectionType == "SQL Dump File" {
		// Get the SQL dump file path
		var filePath string
		filePrompt := &survey.Input{
			Message: "Enter SQL dump file path:",
		}
		if err := survey.AskOne(filePrompt, &filePath); err != nil {
			return config, err
		}

		// Get the original database type
		dbTypePrompt := &survey.Select{
			Message: "Select original database type:",
			Options: []string{"mysql", "postgres", "mariadb"},
		}
		if err := survey.AskOne(dbTypePrompt, &dbTypeStr); err != nil {
			return config, err
		}

		// Parse the SQL dump file
		parser := database.NewSQLDumpParser(filePath, database.DBType(dbTypeStr))
		sqliteDBPath, err := parser.ParseToSQLite()
		if err != nil {
			return config, fmt.Errorf("failed to parse SQL dump file: %w", err)
		}

		// Return SQLite configuration with the temporary database
		return database.Config{
			Type:     database.SQLite,
			FilePath: sqliteDBPath,
		}, nil
	}

	// Get database type for both direct connection and connection string
	dbTypePrompt := &survey.Select{
		Message: "Select database type:",
		Options: []string{"mysql", "postgres", "sqlite3"},
	}
	if err := survey.AskOne(dbTypePrompt, &dbTypeStr); err != nil {
		return config, err
	}

	// Convert string to DBType
	config.Type = database.DBType(dbTypeStr)

	if connectionType == "Connection String" {
		// Get connection string
		var connString string
		connStringPrompt := &survey.Input{
			Message: "Enter connection string:",
			Help:    getConnectionStringHelp(config.Type),
		}
		if err := survey.AskOne(connStringPrompt, &connString); err != nil {
			return config, err
		}
		config.ConnectionURL = connString
		return config, nil
	}

	if config.Type == database.SQLite {
		// SQLite file path
		filePrompt := &survey.Input{
			Message: "Enter SQLite database file path:",
		}
		if err := survey.AskOne(filePrompt, &config.FilePath); err != nil {
			return config, err
		}
	} else {
		// Connection details for MySQL and PostgreSQL
		questions := []*survey.Question{
			{
				Name: "host",
				Prompt: &survey.Input{
					Message: "Enter database host:",
					Default: "localhost",
				},
			},
			{
				Name: "port",
				Prompt: &survey.Input{
					Message: "Enter database port:",
					Default: func() string {
						if config.Type == database.MySQL {
							return "3306"
						}
						return "5432"
					}(),
				},
			},
			{
				Name: "user",
				Prompt: &survey.Input{
					Message: "Enter database user:",
				},
			},
			{
				Name: "password",
				Prompt: &survey.Password{
					Message: "Enter database password:",
				},
			},
			{
				Name: "dbname",
				Prompt: &survey.Input{
					Message: "Enter database name:",
				},
			},
		}

		answers := struct {
			Host     string
			Port     string
			User     string
			Password string
			DBName   string
		}{}

		if err := survey.Ask(questions, &answers); err != nil {
			return config, err
		}

		config.Host = answers.Host
		config.Port = parsePort(answers.Port)
		config.User = answers.User
		config.Password = answers.Password
		config.DBName = answers.DBName
	}

	return config, nil
}

// getConnectionStringHelp returns help text for connection strings based on database type
func getConnectionStringHelp(dbType database.DBType) string {
	switch dbType {
	case database.MySQL:
		return "Format: user:password@tcp(host:port)/dbname\nExample: myuser:mypass@tcp(localhost:3306)/mydb"
	case database.Postgres:
		return "Format: postgresql://user:password@host:port/dbname\nExample: postgresql://myuser:mypass@localhost:5432/mydb"
	case database.SQLite:
		return "Format: path/to/database.db\nExample: ./mydb.sqlite"
	default:
		return ""
	}
}

// SelectTables prompts the user to select tables for export
func SelectTables(db *sql.DB, dbType database.DBType) ([]string, error) {
	// Get tables with row counts
	tableInfos, err := database.GetTablesWithCount(db, dbType)
	if err != nil {
		return nil, err
	}

	if len(tableInfos) == 0 {
		return nil, fmt.Errorf("no tables found in database")
	}

	// Create options with row counts
	var options []string
	tableMap := make(map[string]string) // Maps display string to table name
	for _, info := range tableInfos {
		displayStr := fmt.Sprintf("%s (%d rows)", info.Name, info.RowCount)
		options = append(options, displayStr)
		tableMap[displayStr] = info.Name
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select tables to export:",
		Options: options,
	}

	var selectedDisplay []string
	if err := survey.AskOne(prompt, &selectedDisplay); err != nil {
		return nil, err
	}

	if len(selectedDisplay) == 0 {
		return nil, fmt.Errorf("no tables selected")
	}

	// Convert display strings back to table names
	for _, display := range selectedDisplay {
		if tableName, ok := tableMap[display]; ok {
			selected = append(selected, tableName)
		}
	}

	return selected, nil
}

// SelectOutputDir prompts the user for the output directory
func SelectOutputDir() (string, error) {
	var dir string
	prompt := &survey.Input{
		Message: "Enter output directory for CSV files:",
		Default: ".",
	}

	if err := survey.AskOne(prompt, &dir); err != nil {
		return "", err
	}

	return dir, nil
}

// parsePort converts a string port number to an integer
func parsePort(port string) int {
	var result int
	fmt.Sscanf(port, "%d", &result)
	return result
}
