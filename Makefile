# Makefile pour LockAnalyzer

.PHONY: build clean help test cli

# Variables
BINARY_NAME=lockanalyzer
CLI_BINARY_NAME=lockanalyzer-cli
BUILD_DIR=build

# Compilation
build:
	@echo "🔨 Compilation de l'application principale..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "🔨 Compilation de l'outil CLI..."
	go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME) cmd/lockanalyzer/main.go
	@echo "✅ Compilation terminée"

# Nettoyage
clean:
	@echo "🧹 Nettoyage des fichiers de build..."
	rm -rf $(BUILD_DIR)
	rm -f lock_report_*
	@echo "✅ Nettoyage terminé"

# Aide
help:
	@echo "🔒 LockAnalyzer - Makefile"
	@echo ""
	@echo "Commandes disponibles:"
	@echo "  make build     - Compiler l'application et l'outil CLI"
	@echo "  make clean     - Nettoyer les fichiers de build"
	@echo "  make test      - Lancer les tests"
	@echo "  make cli       - Afficher l'aide de l'outil CLI"
	@echo "  make run       - Lancer l'application principale"
	@echo ""
	@echo "Exemples d'utilisation du CLI:"
	@echo "  ./build/lockanalyzer-cli -help"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -format=markdown"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -format=json -output=report.json"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -interval=10s"

# Tests
test:
	@echo "🧪 Lancement des tests..."
	go test ./...

# Aide CLI
cli: build
	@echo "🔍 Aide de l'outil CLI:"
	@./build/$(CLI_BINARY_NAME) -help

# Lancement de l'application principale
run: build
	@echo "🚀 Lancement de l'application principale..."
	@./build/$(BINARY_NAME)

# Installation (optionnel)
install: build
	@echo "📦 Installation de l'outil CLI..."
	sudo cp build/$(CLI_BINARY_NAME) /usr/local/bin/
	@echo "✅ Installation terminée. Utilisez 'lockanalyzer-cli' depuis n'importe où"

# Désinstallation
uninstall:
	@echo "🗑️  Désinstallation de l'outil CLI..."
	sudo rm -f /usr/local/bin/$(CLI_BINARY_NAME)
	@echo "✅ Désinstallation terminée"

# Exemples d'utilisation
example-markdown: build
	@echo "📝 Exemple: Rapport Markdown vers stdout"
	@./build/$(CLI_BINARY_NAME) -dsn="postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" -format=markdown

example-json: build
	@echo "📊 Exemple: Rapport JSON vers fichier"
	@./build/$(CLI_BINARY_NAME) -dsn="postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" -format=json -output=example_report.json

example-monitoring: build
	@echo "⏰ Exemple: Surveillance en temps réel (5 secondes)"
	@./build/$(CLI_BINARY_NAME) -dsn="postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" -interval=5s 