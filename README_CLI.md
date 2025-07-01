# 🔒 LockAnalyzer CLI

Outil en ligne de commande pour analyser les locks PostgreSQL en temps réel.

## 🚀 Installation

```bash
# Compiler l'outil
make build

# Optionnel : Installer globalement
make install
```

## 📖 Utilisation

### Aide

```bash
./build/lockanalyzer-cli -help
```

### Rapport unique

#### Rapport Markdown vers stdout

```bash
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/db?sslmode=disable" -format=markdown
```

#### Rapport JSON vers fichier

```bash
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/db?sslmode=disable" -format=json -output=report.json
```

#### Rapport texte vers fichier

```bash
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/db?sslmode=disable" -format=text -output=report.txt
```

### Surveillance en temps réel

#### Surveillance vers stdout (toutes les 10 secondes)

```bash
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/db?sslmode=disable" -interval=10s
```

#### Surveillance vers fichiers (toutes les 30 secondes)

```bash
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/db?sslmode=disable" -interval=30s -output=monitoring.md
```

## 🎯 Exemples pratiques

### 1. Analyse rapide d'une base de données

```bash
# Rapport complet en Markdown
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/production?sslmode=disable" -format=markdown
```

### 2. Surveillance pendant un déploiement

```bash
# Surveiller les locks pendant 5 minutes
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/production?sslmode=disable" -interval=15s -output=deployment_monitoring.json
```

### 3. Debug d'un problème de performance

```bash
# Surveillance intensive (toutes les 5 secondes)
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/production?sslmode=disable" -interval=5s -format=text
```

## 📊 Formats de sortie

### Markdown

- **Avantages** : Lisible, structuré, compatible avec les outils de documentation
- **Utilisation** : Rapports pour équipes, documentation, GitHub/GitLab

### JSON

- **Avantages** : Structuré, facilement parsable, intégration avec d'autres outils
- **Utilisation** : Automatisation, monitoring, alertes

### Texte

- **Avantages** : Simple, compatible avec tous les systèmes
- **Utilisation** : Logs, emails, systèmes legacy

## 🔧 Paramètres

| Paramètre   | Type     | Défaut   | Description                             |
| ----------- | -------- | -------- | --------------------------------------- |
| `-dsn`      | string   | -        | DSN PostgreSQL (obligatoire)            |
| `-format`   | string   | markdown | Format de sortie (markdown, json, text) |
| `-output`   | string   | stdout   | Fichier de sortie ou 'stdout'           |
| `-interval` | duration | -        | Intervalle de surveillance (ex: 5s, 1m) |
| `-help`     | bool     | false    | Afficher l'aide                         |

## 🧪 Test avec simulation

Pour tester l'outil avec des locks simulés :

```bash
# Terminal 1 : Lancer la simulation
./scripts/simulate_locks.sh

# Terminal 2 : Surveiller les locks
./build/lockanalyzer-cli -dsn="postgres://user@localhost:5432/testdb?sslmode=disable" -interval=5s
```

## 📈 Métriques analysées

- **Locks actifs** : Nombre et détails des locks PostgreSQL
- **Transactions bloquées** : Transactions en attente de locks
- **Transactions longues** : Transactions exécutées depuis plus de 5 secondes
- **Deadlocks** : Conflits de locks circulaires
- **Conflits d'objets** : Multiples locks sur les mêmes objets
- **Analyse d'index** : Taille et utilisation des index

## 🚨 Suggestions automatiques

L'outil génère automatiquement des suggestions d'amélioration basées sur :

- Présence de transactions bloquées
- Transactions longues
- Conflits d'objets
- Deadlocks détectés
- Nombre élevé de locks

## 🔄 Intégration continue

### Avec Makefile

```bash
# Compilation et test
make build
make test

# Exemples d'utilisation
make example-markdown
make example-json
make example-monitoring
```

### Avec scripts

```bash
# Nettoyage
make clean

# Installation globale
make install
make uninstall
```

## 🛠️ Développement

### Structure du projet

```
concurrent-db/
├── cmd/lockanalyzer/     # Outil CLI
├── formatters/          # Formatters de sortie
├── lockanalyzer/        # Logique d'analyse
├── scripts/             # Scripts utilitaires
├── build/               # Binaires compilés
└── Makefile            # Commandes de build
```

### Ajouter un nouveau format

1. Créer un nouveau formatter dans `formatters/`
2. Implémenter l'interface `LockReportFormatter`
3. Ajouter le cas dans `createFormatter()`
4. Mettre à jour la validation des formats

## 📝 Notes importantes

- **SSL** : Ajouter `?sslmode=disable` au DSN pour les connexions locales
- **Permissions** : L'utilisateur PostgreSQL doit avoir accès aux vues système
- **Performance** : La surveillance en temps réel peut impacter les performances
- **Fichiers** : Les fichiers de sortie sont écrasés s'ils existent déjà
