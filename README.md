# LockAnalyzer 🔒

[![Build Status](https://github.com/pbouamriou/lock-analyzer/workflows/CI/badge.svg)](https://github.com/pbouamriou/lock-analyzer/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pbouamriou/lock-analyzer)](https://golang.org/)
[![License](https://img.shields.io/github/license/pbouamriou/lock-analyzer)](LICENSE)
[![Release](https://img.shields.io/github/v/release/pbouamriou/lock-analyzer)](https://github.com/pbouamriou/lock-analyzer/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pbouamriou/lock-analyzer)](https://goreportcard.com/report/github.com/pbouamriou/lock-analyzer)

A powerful PostgreSQL lock analysis tool written in Go that helps identify and resolve database concurrency issues.

## Features

- 🔍 **Real-time lock monitoring** with configurable intervals
- 📊 **Multiple output formats**: Markdown, JSON, and plain text
- 🌍 **Internationalization** with embedded translation files (French, English, Spanish, German)
- 🚀 **High performance** analysis of large datasets
- 🎯 **Smart suggestions** for lock optimization
- 📈 **Comprehensive reporting** with detailed lock information

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/pbouamriou/lock-analyzer.git
cd lock-analyzer

# Build the application
make build

# Run the CLI tool
./build/lockanalyzer-cli -help
```

### Basic Usage

```bash
# Generate a single report
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=markdown

# Real-time monitoring every 10 seconds
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=10s

# Generate JSON report to file
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=json -output=report.json
```

## Project Structure

```
github.com/pbouamriou/lock-analyzer/
├── cmd/
│   ├── example/           # Example application
│   └── lockanalyzer/      # CLI tool
├── lockanalyzer/          # Core analysis engine
├── formatters/            # Output formatters (Markdown, JSON, Text)
├── i18n/                  # Internationalization
├── locales/               # Embedded translation files
├── scripts/               # Utility scripts
├── testdata/              # Test fixtures
└── Makefile               # Build and test automation
```

## Core Components

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

### Internationalization

Built-in support for multiple languages with embedded translation files:

- French (default)
- English
- Spanish
- German

## Advanced Usage

### Real-time Monitoring

```bash
# Monitor locks every 5 seconds
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=5s

# Monitor with specific language and output format
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=30s -lang=en -format=json -output=monitoring.json
```

### Simulation Script

Use the included simulation script to test lock detection:

```bash
# Start lock simulation
./scripts/simulate_locks.sh

# In another terminal, monitor the locks
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=5s
```

## Configuration

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

## Development

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

### Adding New Languages

1. Create a new translation file in `locales/` (e.g., `es.json`)
2. Add the language to the validation in `i18n/translator.go`
3. Update tests to include the new language
4. Rebuild the application

## API Reference

### CLI Options

| Option      | Description                          | Default           |
| ----------- | ------------------------------------ | ----------------- |
| `-dsn`      | PostgreSQL connection string         | Required          |
| `-format`   | Output format (markdown, json, text) | markdown          |
| `-lang`     | Report language (fr, en, es, de)     | fr                |
| `-output`   | Output file (stdout for screen)      | stdout            |
| `-interval` | Real-time monitoring interval        | 0 (single report) |
| `-help`     | Show help information                | false             |

### Output Formats

#### Markdown

Rich formatting with sections, tables, and emphasis for human reading.

#### JSON

Structured data for programmatic processing and automation.

#### Text

Simple text output suitable for logs and scripts.

## Examples

### Basic Lock Analysis

```bash
# Generate a comprehensive lock report
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=markdown
```

Output:

```markdown
# RAPPORT D'ANALYSE DES LOCKS POSTGRESQL

## RÉSUMÉ EXÉCUTIF

- Total locks actifs: 5
- Transactions bloquées: 2
- Deadlocks détectés: 0
- Problèmes critiques: 1

## LOCKS ACTIFS

- PID: 1234, Mode: ExclusiveLock, Object: users
- PID: 5678, Mode: ShareLock, Object: orders
```

### Continuous Monitoring

```bash
# Monitor locks every 30 seconds
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -interval=30s -lang=en
```

### Integration with CI/CD

```bash
# Generate JSON report for automated analysis
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db" -format=json -output=lock_report.json

# Check for critical issues
if jq '.summary.critical_issues > 0' lock_report.json; then
    echo "Critical lock issues detected!"
    exit 1
fi
```

## Troubleshooting

### Common Issues

1. **Connection refused**: Check PostgreSQL server status and connection parameters
2. **Permission denied**: Ensure the database user has sufficient privileges
3. **No locks detected**: The database might be idle or locks might be too short-lived

### Debug Mode

Enable verbose logging by setting the log level:

```bash
export LOG_LEVEL=debug
./build/lockanalyzer-cli -dsn="postgres://user:pass@localhost:5432/db"
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Add tests for new features
- Update documentation as needed
- Use English for code comments and variable names
- Follow the existing project structure

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Bun ORM](https://bun.uptrace.dev/) for PostgreSQL
- Uses [go-i18n](https://github.com/nicksnyder/go-i18n) for internationalization
- Inspired by PostgreSQL's built-in lock monitoring capabilities

## Support

For questions, issues, or contributions:

- Open an issue on GitHub
- Check the documentation in the `docs/` directory
- Review the test examples for usage patterns
