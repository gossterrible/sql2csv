package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DBType string

const (
	MySQL    DBType = "mysql"
	Postgres DBType = "postgres"
	SQLite   DBType = "sqlite3"
)

type Config struct {
	Type          DBType
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	FilePath      string // For SQLite
	ConnectionURL string // For direct connection string/URL support
}

// Connect establishes a database connection based on the provided configuration
func Connect(config Config) (*sql.DB, error) {
	var dsn string

	if config.ConnectionURL != "" {
		// If a connection URL is provided, use it directly
		dsn = config.ConnectionURL
	} else {
		// Otherwise, build the connection string from individual fields
		switch config.Type {
		case MySQL:
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
				config.User, config.Password, config.Host, config.Port, config.DBName)
		case Postgres:
			dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
				config.Host, config.Port, config.User, config.Password, config.DBName)
		case SQLite:
			dsn = config.FilePath
		default:
			return nil, fmt.Errorf("unsupported database type: %s", config.Type)
		}
	}

	db, err := sql.Open(string(config.Type), dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

// GetTables returns a list of all tables in the database
func GetTables(db *sql.DB, dbType DBType) ([]string, error) {
	var query string

	switch dbType {
	case MySQL:
		query = "SHOW TABLES"
	case Postgres:
		query = `SELECT table_name FROM information_schema.tables 
				WHERE table_schema = 'public'`
	case SQLite:
		query = `SELECT name FROM sqlite_master 
				WHERE type='table' AND name NOT LIKE 'sqlite_%'`
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetColumns returns the column names for a given table
func GetColumns(db *sql.DB, dbType DBType, tableName string) ([]string, error) {
	var query string

	switch dbType {
	case MySQL:
		query = fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)
	case Postgres:
		query = `
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name = $1 
			ORDER BY ordinal_position`
	case SQLite:
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	var columns []string
	var rows *sql.Rows
	var err error

	if dbType == Postgres {
		rows, err = db.Query(query, tableName)
	} else {
		rows, err = db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("error querying columns: %w", err)
	}
	defer rows.Close()

	switch dbType {
	case MySQL:
		for rows.Next() {
			var field, typ, null, key, default_, extra sql.NullString
			if err := rows.Scan(&field, &typ, &null, &key, &default_, &extra); err != nil {
				return nil, err
			}
			columns = append(columns, field.String)
		}
	case Postgres:
		for rows.Next() {
			var column string
			if err := rows.Scan(&column); err != nil {
				return nil, err
			}
			columns = append(columns, column)
		}
	case SQLite:
		for rows.Next() {
			var cid int
			var name, typ, notnull, dfltValue, pk sql.NullString
			if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
				return nil, err
			}
			columns = append(columns, name.String)
		}
	}

	return columns, nil
}

// TableInfo holds table name and its row count
type TableInfo struct {
	Name     string
	RowCount int64
}

// GetTablesWithCount returns a list of all tables in the database with their row counts
func GetTablesWithCount(db *sql.DB, dbType DBType) ([]TableInfo, error) {
	tables, err := GetTables(db, dbType)
	if err != nil {
		return nil, err
	}

	var tableInfos []TableInfo
	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("error counting rows in table %s: %w", table, err)
		}
		tableInfos = append(tableInfos, TableInfo{
			Name:     table,
			RowCount: count,
		})
	}

	return tableInfos, nil
}
