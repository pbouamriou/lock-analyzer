# Makefile pour LockAnalyzer

.PHONY: build clean help test cli

# Variables
BINARY_NAME=lockanalyzer
CLI_BINARY_NAME=lockanalyzer-cli
BUILD_DIR=build
TEST_DIR=testdata

# Cibles principales
.PHONY: all build clean test test-unit test-integration test-coverage run-example run-cli

all: build

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

# Tests unitaires
test-unit:
	@echo "🧪 Lancement des tests unitaires..."
	go test -v ./lockanalyzer/... -run "Test.*" -timeout 30s

# Tests d'intégration
test-integration:
	@echo "🧪 Lancement des tests d'intégration..."
	go test -v ./lockanalyzer/... -run "TestConcurrentTransactions|TestDetectBlockedTransactionsReal|TestGenerateLocksReportWithRealData|TestLockDetectionWithTriggers|TestPerformanceWithLargeDataset" -timeout 60s

# Tests des formatters
test-formatters:
	@echo "🧪 Lancement des tests des formatters..."
	go test -v ./formatters/... -timeout 30s

# Tous les tests
test: test-unit test-formatters test-integration

# Tests avec couverture
test-coverage:
	@echo "🧪 Lancement des tests avec couverture..."
	go test -v -coverprofile=coverage.out ./lockanalyzer/... ./formatters/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "🎯 Rapport de couverture généré: coverage.html"

# Exécution de l'exemple
run-example:
	@echo "🚀 Lancement de l'exemple..."
	go run ./cmd/example/main.go

# Exécution du CLI
run-cli:
	@echo "🔍 Lancement du CLI..."
	@if [ ! -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then make build; fi
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Simulation de locks
simulate-locks:
	@echo "🔄 Simulation de locks..."
	@chmod +x scripts/simulate_locks.sh
	./scripts/simulate_locks.sh

# Génération de rapports de test
test-reports:
	@echo "🎯 Génération de rapports de test..."
	@if [ ! -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then make build; fi
	./$(BUILD_DIR)/$(BINARY_NAME) --dsn "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" --format markdown --output test_report.md
	./$(BUILD_DIR)/$(BINARY_NAME) --dsn "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" --format json --output test_report.json
	./$(BUILD_DIR)/$(BINARY_NAME) --dsn "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" --format text --output test_report.txt

# Installation des dépendances
deps:
	@echo "📦 Installation des dépendances..."
	go mod download
	go mod tidy

# Vérification du code
lint:
	@echo "🔍 Vérification du code..."
	gofmt -s -w .
	golint ./...
	go vet ./...

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
	@echo "  make test-unit      - Exécuter les tests unitaires"
	@echo "  make test-integration - Exécuter les tests d'intégration"
	@echo "  make test-formatters - Exécuter les tests des formatters"
	@echo "  make test-coverage  - Exécuter les tests avec couverture"
	@echo "  make run-example    - Exécuter l'exemple"
	@echo "  make run-cli        - Exécuter le CLI"
	@echo "  make simulate-locks - Simuler des locks"
	@echo "  make test-reports   - Générer des rapports de test"
	@echo "  make deps           - Installer les dépendances"
	@echo "  make lint           - Vérifier le code"
	@echo "  make help           - Afficher cette aide"
	@echo ""
	@echo "Exemples d'utilisation du CLI:"
	@echo "  ./build/lockanalyzer-cli -help"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -format=markdown"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -format=json -output=report.json"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -interval=10s"

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