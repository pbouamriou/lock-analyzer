#!/bin/bash

# Script pour simuler des locks PostgreSQL
# Utilise l'application principale pour créer des locks

echo "🔒 Simulation de locks PostgreSQL"
echo "=================================="

# Vérifier que l'application est compilée
if [ ! -f "build/lockanalyzer" ]; then
    echo "❌ L'application n'est pas compilée. Lancez 'make build' d'abord."
    exit 1
fi

echo "🚀 Lancement de l'application de test (Ctrl+C pour arrêter)..."
echo "📊 Utilisez l'outil CLI pour surveiller les locks:"
echo "   ./build/lockanalyzer-cli -dsn='postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable' -interval=5s"
echo ""

# Lancer l'application principale
./build/lockanalyzer 