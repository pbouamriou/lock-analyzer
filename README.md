# LockAnalyzer 🔒

[![Build Status](https://github.com/pbouamriou/lock-analyzer/workflows/CI/badge.svg)](https://github.com/pbouamriou/lock-analyzer/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pbouamriou/lock-analyzer)](https://golang.org/)
[![License](https://img.shields.io/github/license/pbouamriou/lock-analyzer)](LICENSE)
[![Release](https://img.shields.io/github/v/release/pbouamriou/lock-analyzer)](https://github.com/pbouamriou/lock-analyzer/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pbouamriou/lock-analyzer)](https://goreportcard.com/report/github.com/pbouamriou/lock-analyzer)

A powerful PostgreSQL lock analysis tool written in Go that helps identify and resolve database concurrency issues.

## 🚀 Features

- 🔍 **Real-time lock monitoring** with configurable intervals
- 📊 **Multiple output formats**: Markdown, JSON, and plain text
- 🌍 **Internationalization** with embedded translation files (French, English, Spanish, German)
- 🚀 **High performance** analysis of large datasets
- 🎯 **Smart suggestions** for lock optimization
- 📈 **Comprehensive reporting** with detailed lock information

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/pbouamriou/lock-analyzer.git
cd lock-analyzer

# Build the application
make build

# Optional: Install globally
make install
```

## 🎯 Quick Start

### Help

```bash
./build/lockanalyzer-cli -help
```

### Single Report

#### Markdown report to stdout

```bash
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=markdown
```

#### JSON report to file

```bash
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=json -output=report.json
```

#### Text report to file

```bash
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=text -output=report.txt
```

### Real-time Monitoring

#### Monitoring to stdout (every 10 seconds)

```bash
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=10s
```

#### Monitoring to files (every 30 seconds)

```bash
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=30s -output=monitoring.md
```

## 🎯 Practical Examples

### 1. Quick database analysis

```bash
# Complete Markdown report
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/production?sslmode=disable" -format=markdown
```

### 2. Monitoring during deployment

```bash
# Monitor locks for 5 minutes
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/production?sslmode=disable" -interval=15s -output=deployment_monitoring.json
```

### 3. Performance issue debugging

```bash
# Intensive monitoring (every 5 seconds)
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/production?sslmode=disable" -interval=5s -format=text
```

### 4. Monitoring with specific language and format

```bash
# Monitor with English language and JSON format
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=30s -lang=en -format=json -output=monitoring.json
```

## 📊 Output Formats

### Markdown

- **Advantages**: Readable, structured, compatible with documentation tools
- **Usage**: Reports for teams, documentation, GitHub/GitLab

### JSON

- **Advantages**: Structured, easily parsable, integration with other tools
- **Usage**: Automation, monitoring, alerts

### Text

- **Advantages**: Simple, compatible with all systems
- **Usage**: Logs, emails, legacy systems

## 🔧 Configuration

### CLI Parameters

| Parameter   | Type     | Default  | Description                             |
| ----------- | -------- | -------- | --------------------------------------- |
| `-dsn`      | string   | -        | PostgreSQL connection string (required) |
| `-format`   | string   | markdown | Output format (markdown, json, text)    |
| `-lang`     | string   | fr       | Report language (fr, en, es, de)        |
| `-output`   | string   | stdout   | Output file or 'stdout'                 |
| `-interval` | duration | -        | Monitoring interval (e.g., 5s, 1m)      |
| `-help`     | bool     | false    | Show help                               |

### Database Connection

The tool uses standard PostgreSQL connection strings:

```bash
# Basic connection
postgres://user:pass@localhost:5432/db

# With SSL
postgres://user:pass@localhost:5432/db?sslmode=require

# With additional parameters
postgres://user:pass@localhost:5432/db?sslmode=disable&connect_timeout=10
```

### Environment Variables

- `LANG`: System language detection
- `LC_ALL`: Alternative language setting
- `LC_MESSAGES`: Message language preference

## 🧪 Testing with Simulation

To test the tool with simulated locks:

```bash
# Terminal 1: Start simulation
./scripts/simulate_locks.sh

# Terminal 2: Monitor locks
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/testdb?sslmode=disable" -interval=5s
```

## 📈 Analyzed Metrics

- **Active locks**: Number and details of PostgreSQL locks
- **Blocked transactions**: Transactions waiting for locks
- **Long transactions**: Transactions running for more than 5 seconds
- **Deadlocks**: Circular lock conflicts
- **Object conflicts**: Multiple locks on the same objects
- **Index analysis**: Index size and usage

## 🚨 Automatic Suggestions

The tool automatically generates improvement suggestions based on:

- Presence of blocked transactions
- Long transactions
- Object conflicts
- Detected deadlocks
- High number of locks

## 🌍 Internationalization

### Embedded Translation Files

This project uses Go's embedded file system (embed) to include translation files directly in the binary, avoiding missing file issues during installation.

#### Architecture

```
github.com/pbouamriou/lock-analyzer/
├── locales/               # Embedded translation files
│   ├── en.json           # English translations
│   ├── fr.json           # French translations (default)
│   ├── es.json           # Spanish translations
│   ├── de.json           # German translations
│   ├── embedded.go       # Embedded file system implementation
│   └── embedded_test.go  # Embedded file tests
├── i18n/                  # Internationalization system
│   ├── translator.go      # Translation manager and language detection
│   └── translator_test.go # Translation tests
└── cmd/
    └── lockanalyzer/
        └── main.go        # CLI entry point with i18n initialization
```

#### Advantages of Embedded Files

1. **Portability**: The binary contains all translations
2. **Installation simplicity**: No external files required
3. **Consistency**: Translations are always available
4. **Performance**: Fast loading from memory
5. **Security**: No external manipulation of translation files

#### Usage

```bash
# Tool works immediately without external files
./build/lockanalyzer-cli -help

# Language change
./build/lockanalyzer-cli -help -lang=en
./build/lockanalyzer-cli -help -lang=fr
```

#### Adding New Languages

1. **Create translation file**:

   ```bash
   # Create locales/es.json for Spanish
   cp locales/fr.json locales/es.json
   # Modify translations in locales/es.json
   ```

2. **Update tests**:

   ```go
   // In tests, add the new language
   expectedFiles := map[string]bool{
       "en.json": false,
       "fr.json": false,
       "es.json": false,  // New language
   }
   ```

3. **Rebuild**:
   ```bash
   make build
   ```

## 🏗️ Project Structure

```
github.com/pbouamriou/lock-analyzer/
├── cmd/
│   ├── example/           # Example application with usage patterns
│   │   └── main.go        # Example implementation
│   └── lockanalyzer/      # CLI tool entry point
│       ├── main.go        # Main CLI application
│       └── main_test.go   # CLI tests
├── lockanalyzer/          # Core analysis engine
│   ├── lockanalyzer.go    # Main analysis logic and PostgreSQL queries
│   ├── lockanalyzer_test.go # Core engine tests
│   ├── integration_test.go # Integration tests
│   └── test_utils.go      # Test utilities and helpers
├── formatters/            # Output formatters (Markdown, JSON, Text)
│   ├── formatters.go      # Formatter interface and factory
│   ├── markdown.go        # Markdown formatter implementation
│   ├── json.go           # JSON formatter implementation
│   ├── text.go           # Text formatter implementation
│   ├── templates.go      # Template management
│   ├── templates/        # Template files
│   │   ├── markdown.tmpl # Markdown template
│   │   └── text.tmpl     # Text template
│   ├── example_usage.go  # Formatter usage examples
│   ├── formatters_test.go # Formatter tests
│   └── formatters_i18n_test.go # Internationalization tests
├── i18n/                  # Internationalization system
│   ├── translator.go      # Translation manager and language detection
│   └── translator_test.go # Translation tests
├── locales/               # Embedded translation files
│   ├── en.json           # English translations
│   ├── fr.json           # French translations (default)
│   ├── es.json           # Spanish translations
│   ├── de.json           # German translations
│   ├── embedded.go       # Embedded file system implementation
│   └── embedded_test.go  # Embedded file tests
├── scripts/               # Utility scripts
│   └── simulate_locks.sh # PostgreSQL lock simulation script
├── testdata/              # Test fixtures and data
│   ├── fixture_example.yml # Example test data
│   └── fixture_test.yml  # Test fixtures
├── docs/                  # Documentation
│   └── badges.md         # Badge documentation
├── database/              # Database utilities (future use)
├── db-model/              # Database models (future use)
├── config/                # Configuration management (future use)
├── embedded/              # Embedded resources (future use)
├── assets/                # Static assets (future use)
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile               # Build and test automation
├── LICENSE                # MIT License
└── README.md              # This documentation
```

## 🔧 Core Components

### LockAnalyzer Engine

The core analysis engine provides:

- **Lock Detection**: Identifies active locks and their types
- **Transaction Analysis**: Detects blocked and long-running transactions
- **Conflict Resolution**: Analyzes object conflicts and deadlocks
- **Performance Insights**: Index analysis and optimization suggestions

### Formatters

Multiple output formats for different use cases:

- **Markdown**: Human-readable reports with rich formatting
- **JSON**: Machine-readable output for automation
- **Text**: Simple text output for logs and scripts

## 🚀 Development

### Building

```bash
# Build all components
make build

# Build specific component
go build -o build/lockanalyzer-cli cmd/lockanalyzer/main.go
```

### Testing

```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration
make test-formatters

# Run tests with coverage
make test-coverage
```

### Continuous Integration

```bash
# Build and test
make build
make test

# Usage examples
make example-markdown
make example-json
make example-monitoring

# Cleanup
make clean

# Global installation
make install
make uninstall
```

### Adding a New Format

1. Create a new formatter in `formatters/`
2. Implement the `LockReportFormatter` interface
3. Add the case in `createFormatter()`
4. Update format validation

## 📝 Important Notes

- **SSL**: Add `?sslmode=disable` to DSN for local connections
- **Permissions**: PostgreSQL user must have access to system views
- **Performance**: Real-time monitoring may impact performance
- **Files**: Output files are overwritten if they already exist

## 🔄 Migration from Old Approach

### Before (External Files)

```go
// Search in multiple directories
possiblePaths := []string{
    "locales",
    "../locales",
    "../../locales",
}
```

### After (Embedded Files)

```go
// Direct access to embedded files
localesFS := locales.GetLocalesFS()
content, err := fs.ReadFile(localesFS, "fr.json")
```

## 🛠️ Maintenance

### Verifying Embedded Files

```bash
# List embedded files
go run -c 'package main; import "github.com/pbouamriou/lock-analyzer/locales"; func main() { files, _ := locales.ListLocaleFiles(); for _, f := range files { println(f) } }'

# Check file content
go run -c 'package main; import "github.com/pbouamriou/lock-analyzer/locales"; func main() { content, _ := locales.GetLocaleFile("fr.json"); println(string(content)) }'

# Alternative: Use the built tool
./build/lockanalyzer-cli -help -lang=en
./build/lockanalyzer-cli -help -lang=fr
```

### Updating Translations

1. Modify JSON files in `locales/`
2. Rebuild the application
3. New translations are automatically included

## 📚 API Reference

The API reference provides detailed information about the LockAnalyzer components and interfaces. For CLI parameters, see the [Configuration](#-configuration) section above.

### Core Interfaces

#### LockReportFormatter

```go
type LockReportFormatter interface {
    Format(report *LockReport) (string, error)
}
```

#### LockAnalyzer

```go
type LockAnalyzer struct {
    db *sql.DB
}

func (la *LockAnalyzer) Analyze() (*LockReport, error)
func (la *LockAnalyzer) Monitor(interval time.Duration, output chan<- *LockReport)
```

### Data Structures

#### LockReport

```go
type LockReport struct {
    Summary       SummaryInfo
    ActiveLocks   []LockInfo
    BlockedTxns   []TransactionInfo
    LongTxns      []TransactionInfo
    Suggestions   []string
    GeneratedAt   time.Time
}
```

### Internationalization

#### Translator

```go
type Translator struct {
    bundle *i18n.Bundle
}

func (t *Translator) Translate(key string, lang string) string
func (t *Translator) GetSupportedLanguages() []string
```

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
