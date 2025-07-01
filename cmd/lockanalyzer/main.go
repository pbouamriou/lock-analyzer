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
	// Configuration des flags
	var (
		dsn      = flag.String("dsn", "", "DSN de connexion PostgreSQL (ex: postgres://user:pass@localhost:5432/db)")
		format   = flag.String("format", "markdown", "Format de sortie: markdown, json, text")
		lang     = flag.String("lang", "fr", "Langue du rapport: fr, en, es, de")
		output   = flag.String("output", "stdout", "Fichier de sortie ou 'stdout' pour l'affichage")
		interval = flag.Duration("interval", 0, "Intervalle de surveillance en temps réel (ex: 5s, 1m)")
		help     = flag.Bool("help", false, "Afficher l'aide")
	)
	flag.Parse()

	// Affichage de l'aide
	if *help {
		printHelp()
		return
	}

	// Validation des paramètres
	if *dsn == "" {
		log.Fatal("Le paramètre -dsn est obligatoire")
	}

	// Validation du format
	validFormats := map[string]bool{"markdown": true, "json": true, "text": true}
	if !validFormats[*format] {
		log.Fatalf("Format invalide: %s. Formats supportés: markdown, json, text", *format)
	}

	// Validation de la langue
	validLangs := map[string]bool{"fr": true, "en": true, "es": true, "de": true}
	if !validLangs[*lang] {
		log.Fatalf("Langue invalide: %s. Langues supportées: fr, en, es, de", *lang)
	}

	// Connexion à la base de données
	db, err := connectDB(*dsn)
	if err != nil {
		log.Fatalf("Erreur de connexion à la base de données: %v", err)
	}
	defer db.Close()

	// Création du formatter
	formatter, err := formatters.NewFormatter(*format, *lang)
	if err != nil {
		log.Fatalf("Erreur lors de la création du formatter: %v", err)
	}

	// Mode surveillance en temps réel
	if *interval > 0 {
		runRealTimeMonitoring(db, formatter, *interval, *output)
		return
	}

	// Mode rapport unique
	generateSingleReport(db, formatter, *output)
}

func printHelp() {
	fmt.Print(`🔒 LockAnalyzer - Outil d'analyse des locks PostgreSQL

USAGE:
  lockanalyzer -dsn="postgres://user:pass@localhost:5432/db" [options]

PARAMÈTRES OBLIGATOIRES:
  -dsn string
        DSN de connexion PostgreSQL
        Exemple: postgres://user:pass@localhost:5432/db

OPTIONS:
  -format string
        Format de sortie (défaut: markdown)
        Valeurs: markdown, json, text

  -lang string
        Langue du rapport (défaut: fr)
        Valeurs: fr, en, es, de

  -output string
        Fichier de sortie (défaut: stdout)
        Utiliser 'stdout' pour l'affichage à l'écran

  -interval duration
        Surveillance en temps réel
        Exemples: 5s, 30s, 1m, 5m

  -help
        Afficher cette aide

EXEMPLES:
  # Rapport unique en Markdown vers stdout (français)
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -format=markdown

  # Rapport JSON en anglais vers fichier
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -format=json -lang=en -output=report.json

  # Surveillance en temps réel toutes les 10 secondes (espagnol)
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -interval=10s -lang=es

  # Surveillance en temps réel vers fichier (allemand)
  lockanalyzer -dsn="postgres://user@localhost:5432/testdb" -interval=30s -lang=de -output=monitoring.txt
`)
}

func connectDB(dsn string) (*bun.DB, error) {
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test de la connexion
	if err := sqldb.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la base de données: %v", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	return db, nil
}

func generateSingleReport(db *bun.DB, formatter formatters.LockReportFormatter, output string) {
	fmt.Printf("🔍 Génération du rapport d'analyse des locks...\n")

	if output == "stdout" {
		// Affichage vers stdout
		if err := formatters.GenerateAndDisplayReport(db, formatter); err != nil {
			log.Fatalf("Erreur lors de la génération du rapport: %v", err)
		}
	} else {
		// Écriture vers fichier
		if err := formatters.GenerateAndWriteReport(db, formatter, output); err != nil {
			log.Fatalf("Erreur lors de l'écriture du rapport: %v", err)
		}
		fmt.Printf("✅ Rapport généré: %s\n", output)
	}
}

func runRealTimeMonitoring(db *bun.DB, formatter formatters.LockReportFormatter, interval time.Duration, output string) {
	fmt.Printf("🔍 Surveillance en temps réel des locks (intervalle: %s)\n", interval)
	fmt.Printf("📁 Sortie: %s\n", output)
	fmt.Printf("⏹️  Appuyez sur Ctrl+C pour arrêter\n\n")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Gestion de l'interruption
	sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			timestamp := time.Now().Format("15:04:05")

			if output == "stdout" {
				fmt.Printf("\n--- Analyse #%d (%s) ---\n", counter, timestamp)
				if err := formatters.GenerateAndDisplayReport(db, formatter); err != nil {
					log.Printf("Erreur lors de la génération du rapport: %v", err)
				}
			} else {
				// Générer un nom de fichier avec timestamp
				ext := formatter.GetFileExtension()
				filename := fmt.Sprintf("%s_%s_%03d%s",
					strings.TrimSuffix(output, ext),
					time.Now().Format("20060102_150405"),
					counter,
					ext)

				if err := formatters.GenerateAndWriteReport(db, formatter, filename); err != nil {
					log.Printf("Erreur lors de l'écriture du rapport: %v", err)
				} else {
					fmt.Printf("✅ Rapport #%d généré: %s\n", counter, filename)
				}
			}

		case <-sigChan:
			fmt.Println("\n🛑 Surveillance arrêtée par l'utilisateur")
			return
		}
	}
}
