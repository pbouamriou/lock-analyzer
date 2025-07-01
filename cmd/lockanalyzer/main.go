package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"concurrent-db/formatters"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func main() {
	// Flag configuration
	var (
		dsn      = flag.String("dsn", "", "PostgreSQL connection DSN (e.g., postgres://user:pass@localhost:5432/db)")
		format   = flag.String("format", "markdown", "Output format: markdown, json, text")
		lang     = flag.String("lang", "fr", "Report language: fr, en, es, de")
		output   = flag.String("output", "stdout", "Output file or 'stdout' for display")
		interval = flag.Duration("interval", 0, "Real-time monitoring interval (e.g., 5s, 1m)")
		help     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	// Display help
	if *help {
		printHelp()
		return
	}

	// Parameter validation
	if *dsn == "" {
		log.Fatal("The -dsn parameter is required")
	}

	// Format validation
	validFormats := map[string]bool{"markdown": true, "json": true, "text": true}
	if !validFormats[*format] {
		log.Fatalf("Invalid format: %s. Supported formats: markdown, json, text", *format)
	}

	// Language validation
	validLangs := map[string]bool{"fr": true, "en": true, "es": true, "de": true}
	if !validLangs[*lang] {
		log.Fatalf("Invalid language: %s. Supported languages: fr, en, es, de", *lang)
	}

	// Database connection
	db, err := connectDB(*dsn)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Create formatter
	formatter, err := formatters.NewFormatter(*format, *lang)
	if err != nil {
		log.Fatalf("Error creating formatter: %v", err)
	}

	// Real-time monitoring mode
	if *interval > 0 {
		runRealTimeMonitoring(db, formatter, *interval, *output)
		return
	}

	// Single report mode
	generateSingleReport(db, formatter, *output)
}

func printHelp() {
	fmt.Print(`🔒 LockAnalyzer - PostgreSQL Lock Analysis Tool

USAGE:
  lockanalyzer -dsn="postgres://user:pass@localhost:5432/db" [options]

REQUIRED PARAMETERS:
  -dsn string
        PostgreSQL connection DSN
        Example: postgres://user:pass@localhost:5432/db

OPTIONS:
  -format string
        Output format (default: markdown)
        Values: markdown, json, text

  -lang string
        Report language (default: fr)
        Values: fr, en, es, de

  -output string
        Output file (default: stdout)
        Use 'stdout' for screen display

  -interval duration
        Real-time monitoring
        Examples: 5s, 30s, 1m, 5m

  -help
        Show this help

EXAMPLES:
  # Single report in Markdown to stdout (French)
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -format=markdown

  # JSON report in English to file
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -format=json -lang=en -output=report.json

  # Real-time monitoring every 10 seconds (Spanish)
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -interval=10s -lang=es

  # Real-time monitoring to file (German)
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -interval=30s -lang=de -output=monitoring.txt
`)
}

func connectDB(dsn string) (*bun.DB, error) {
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := sqldb.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	return db, nil
}

func generateSingleReport(db *bun.DB, formatter formatters.LockReportFormatter, output string) {
	fmt.Printf("🔍 Generating lock analysis report...\n")

	if output == "stdout" {
		// Display to stdout
		if err := formatters.GenerateAndDisplayReport(db, formatter); err != nil {
			log.Fatalf("Error generating report: %v", err)
		}
	} else {
		// Write to file
		if err := formatters.GenerateAndWriteReport(db, formatter, output); err != nil {
			log.Fatalf("Error writing report: %v", err)
		}
		fmt.Printf("✅ Report generated: %s\n", output)
	}
}

func runRealTimeMonitoring(db *bun.DB, formatter formatters.LockReportFormatter, interval time.Duration, output string) {
	fmt.Printf("🔍 Real-time lock monitoring (interval: %s)\n", interval)
	fmt.Printf("📁 Output: %s\n", output)
	fmt.Printf("⏹️  Press Ctrl+C to stop\n\n")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Interrupt handling
	sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			timestamp := time.Now().Format("15:04:05")

			if output == "stdout" {
				fmt.Printf("\n--- Analysis #%d (%s) ---\n", counter, timestamp)
				if err := formatters.GenerateAndDisplayReport(db, formatter); err != nil {
					log.Printf("Error generating report: %v", err)
				}
			} else {
				// Generate filename with timestamp
				ext := formatter.GetFileExtension()
				filename := fmt.Sprintf("%s_%s_%03d%s",
					strings.TrimSuffix(output, ext),
					time.Now().Format("20060102_150405"),
					counter,
					ext)

				if err := formatters.GenerateAndWriteReport(db, formatter, filename); err != nil {
					log.Printf("Error writing report: %v", err)
				} else {
					fmt.Printf("✅ Report #%d generated: %s\n", counter, filename)
				}
			}

		case <-sigChan:
			fmt.Println("\n🛑 Monitoring stopped by user")
			return
		}
	}
}
