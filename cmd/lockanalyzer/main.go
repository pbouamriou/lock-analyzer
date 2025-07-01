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
	"concurrent-db/i18n"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func main() {
	// Get language from environment or default to French
	lang := getLanguageFromEnv()

	// Create translator for CLI messages
	translator := i18n.NewTranslator(lang)

	// Flag configuration with localized descriptions
	var (
		dsn      = flag.String("dsn", "", translator.T("cli_dsn_description"))
		format   = flag.String("format", "markdown", translator.T("cli_format_description"))
		langFlag = flag.String("lang", lang, translator.T("cli_lang_description"))
		output   = flag.String("output", "stdout", translator.T("cli_output_description"))
		interval = flag.Duration("interval", 0, translator.T("cli_interval_description"))
		help     = flag.Bool("help", false, translator.T("cli_help_description"))
	)
	flag.Parse()

	// Update translator if language was specified via flag
	if *langFlag != lang {
		translator = i18n.NewTranslator(*langFlag)
	}

	// Display help
	if *help {
		printHelp(translator)
		return
	}

	// Parameter validation
	if *dsn == "" {
		log.Fatal(translator.T("cli_dsn_required"))
	}

	// Format validation
	validFormats := map[string]bool{"markdown": true, "json": true, "text": true}
	if !validFormats[*format] {
		log.Fatalf(translator.T("cli_invalid_format"), *format)
	}

	// Language validation
	validLangs := map[string]bool{"fr": true, "en": true, "es": true, "de": true}
	if !validLangs[*langFlag] {
		log.Fatalf(translator.T("cli_invalid_language"), *langFlag)
	}

	// Database connection
	db, err := connectDB(*dsn)
	if err != nil {
		log.Fatalf(translator.T("cli_db_connection_error"), err)
	}
	defer db.Close()

	// Create formatter
	formatter, err := formatters.NewFormatter(*format, *langFlag)
	if err != nil {
		log.Fatalf(translator.T("cli_formatter_error"), err)
	}

	// Real-time monitoring mode
	if *interval > 0 {
		runRealTimeMonitoring(db, formatter, *interval, *output, translator)
		return
	}

	// Single report mode
	generateSingleReport(db, formatter, *output, translator)
}

// getLanguageFromEnv detects the system language from environment variables
func getLanguageFromEnv() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}

	// Extract language code
	if strings.Contains(lang, "fr") {
		return "fr"
	}
	if strings.Contains(lang, "en") {
		return "en"
	}
	if strings.Contains(lang, "es") {
		return "es"
	}
	if strings.Contains(lang, "de") {
		return "de"
	}

	return "fr" // Default to French
}

func printHelp(translator *i18n.Translator) {
	fmt.Printf(`🔒 %s

%s:
  lockanalyzer -dsn="postgres://user:pass@localhost:5432/db" [options]

%s:
  -dsn string
        %s
        %s

%s:
  -format string
        %s
        %s

  -lang string
        %s
        %s

  -output string
        %s
        %s

  -interval duration
        %s
        %s

  -help
        %s

%s:
  # %s
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -format=markdown

  # %s
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -format=json -lang=en -output=report.json

  # %s
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -interval=10s -lang=es

  # %s
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -interval=30s -lang=de -output=monitoring.txt
`,
		translator.T("cli_tool_title"),
		translator.T("cli_usage"),
		translator.T("cli_required_parameters"),
		translator.T("cli_dsn_description"),
		translator.T("cli_dsn_example"),
		translator.T("cli_options"),
		translator.T("cli_format_description"),
		translator.T("cli_format_values"),
		translator.T("cli_lang_description"),
		translator.T("cli_lang_values"),
		translator.T("cli_output_description"),
		translator.T("cli_output_use"),
		translator.T("cli_interval_description"),
		translator.T("cli_interval_examples"),
		translator.T("cli_help_description"),
		translator.T("cli_examples"),
		translator.T("cli_example_1"),
		translator.T("cli_example_2"),
		translator.T("cli_example_3"),
		translator.T("cli_example_4"),
	)
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

func generateSingleReport(db *bun.DB, formatter formatters.LockReportFormatter, output string, translator *i18n.Translator) {
	fmt.Printf("🔍 %s\n", translator.T("cli_generating_report"))

	if output == "stdout" {
		// Display to stdout
		if err := formatters.GenerateAndDisplayReport(db, formatter); err != nil {
			log.Fatalf(translator.T("cli_report_generation_error"), err)
		}
	} else {
		// Write to file
		if err := formatters.GenerateAndWriteReport(db, formatter, output); err != nil {
			log.Fatalf(translator.T("cli_report_writing_error"), err)
		}
		fmt.Printf("✅ %s: %s\n", translator.T("cli_report_generated"), output)
	}
}

func runRealTimeMonitoring(db *bun.DB, formatter formatters.LockReportFormatter, interval time.Duration, output string, translator *i18n.Translator) {
	fmt.Printf("🔍 %s\n", fmt.Sprintf(translator.T("cli_realtime_monitoring"), interval))
	fmt.Printf("📁 %s: %s\n", translator.T("cli_output"), output)
	fmt.Printf("⏹️  %s\n\n", translator.T("cli_press_ctrl_c"))

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
				fmt.Printf("\n--- %s #%d (%s) ---\n", translator.T("cli_analysis"), counter, timestamp)
				if err := formatters.GenerateAndDisplayReport(db, formatter); err != nil {
					log.Printf(translator.T("cli_report_generation_error"), err)
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
					log.Printf(translator.T("cli_report_writing_error"), err)
				} else {
					fmt.Printf("✅ %s #%d: %s\n", translator.T("cli_report_generated"), counter, filename)
				}
			}

		case <-sigChan:
			fmt.Printf("\n🛑 %s\n", translator.T("cli_monitoring_stopped"))
			return
		}
	}
}
