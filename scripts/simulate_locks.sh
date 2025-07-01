#!/bin/bash

# Script to simulate PostgreSQL locks
# Uses the main application to create locks

echo "🔒 PostgreSQL Lock Simulation"
echo "============================="

# Check if the application is compiled
if [ ! -f "build/lockanalyzer-example" ]; then
    echo "❌ Application is not compiled. Run 'make build' first."
    exit 1
fi

echo "🚀 Starting test application (Ctrl+C to stop)..."
echo "📊 Use the CLI tool to monitor locks:"
echo "   ./build/lockanalyzer-cli -dsn='postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable' -interval=5s"
echo ""

# Launch the example application
./build/lockanalyzer-example 