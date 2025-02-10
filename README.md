# sql2csv

[![Go Report Card](https://goreportcard.com/badge/github.com/gossterrible/sql2csv)](https://goreportcard.com/report/github.com/gossterrible/sql2csv)
[![GoDoc](https://godoc.org/github.com/gossterrible/sql2csv?status.svg)](https://godoc.org/github.com/gossterrible/sql2csv)
[![Release](https://img.shields.io/github/release/gossterrible/sql2csv.svg)](https://github.com/gossterrible/sql2csv/releases/latest)
[![License](https://img.shields.io/github/license/gossterrible/sql2csv.svg)](LICENSE)

A powerful command-line tool to export SQL database tables to CSV files. Supports MySQL, PostgreSQL, and SQLite databases with an interactive interface and concurrent exports.

## Features

- 🗃️ Support for multiple SQL database types:
  - MySQL/MariaDB
  - PostgreSQL
  - SQLite
- 🔌 Multiple connection methods:
  - Interactive connection details input
  - Connection string/URL support
  - SQL dump file import
- 📊 Interactive table selection with row count display
- ⚡ Concurrent export of multiple tables
- 🚀 Efficient handling of large tables through batch processing
- 🛠️ User-friendly command-line interface
- 🔒 Secure password handling

## Installation

### Using Go Install

```bash
go install github.com/gossterrible/sql2csv/cmd/sql2csv@latest
```

### From Release Binary

Download the latest binary for your platform from the [releases page](https://github.com/gossterrible/sql2csv/releases).

### Building from Source

```bash
# Clone the repository
git clone https://github.com/gossterrible/sql2csv.git
cd sql2csv

# Install dependencies
go mod download

# Build the binary
go build ./cmd/sql2csv

# Run tests
go test ./...
```

## Usage

### Quick Start

Simply run the tool and follow the interactive prompts:

```bash
sql2csv
```

### Connection Methods

#### 1. Direct Connection
```bash
$ sql2csv
? Select connection type: Direct Connection
? Select database type: postgres
? Enter database host: localhost
? Enter database port: 5432
? Enter database user: myuser
? Enter database password: ****
? Enter database name: mydb
```

#### 2. Connection String
```bash
$ sql2csv
? Select connection type: Connection String
? Select database type: mysql
? Enter connection string: myuser:mypass@tcp(localhost:3306)/mydb
```

#### 3. SQL Dump File
```bash
$ sql2csv
? Select connection type: SQL Dump File
? Enter SQL dump file path: ./dump.sql
? Select original database type: postgres
```

### Example Output Structure

```
output_directory/
├── users.csv
├── products.csv
└── orders.csv
```

## Database Support Details

### MySQL/MariaDB
- Supports all MySQL data types
- Default port: 3306
- Connection string format: `user:password@tcp(host:port)/dbname`
- Required permissions: SELECT on target tables

### PostgreSQL
- Supports all PostgreSQL data types
- Default port: 5432
- Connection string format: `postgresql://user:password@host:port/dbname`
- SSL mode disabled by default
- Required permissions: SELECT on target tables and schema information

### SQLite
- Supports all SQLite data types
- Connection format: Path to database file
- No additional server setup needed
- Read permissions required on the database file

## Performance Optimization

The tool implements several optimizations for handling large datasets:

- Batch processing to minimize memory usage
- Concurrent table exports using Go routines
- Efficient CSV writing with buffering
- Connection pooling for better resource utilization

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Requirements

- Go 1.16 or later
- CGO enabled (required for SQLite support)
- Access to test databases (MySQL, PostgreSQL, SQLite) for integration testing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) - MySQL driver
- [lib/pq](https://github.com/lib/pq) - PostgreSQL driver
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [survey](https://github.com/AlecAivazis/survey) - Interactive prompts 