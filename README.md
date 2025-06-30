# Concurrent Database Test

Ce projet teste les comportements de concurrence dans PostgreSQL en utilisant une configuration YAML pour initialiser la base de données.

## Structure du projet

```
concurrent-db/
├── config.yaml          # Configuration de la base de données
├── config/
│   └── config.go        # Gestion de la configuration YAML
├── database/
│   └── init.go          # Initialisation de la base de données
├── main.go              # Programme principal
└── README.md           # Ce fichier
```

## Configuration

Le fichier `config.yaml` définit :

- **Database** : Paramètres de connexion PostgreSQL
- **Tables** : Structure des tables avec colonnes et contraintes
- **SampleData** : Données d'exemple à insérer

### Exemple de configuration

```yaml
database:
  host: localhost
  port: 5432
  user: philippebouamriou
  password: ""
  name: testdb
  sslmode: disable

tables:
  - name: projects
    columns:
      - name: id
        type: uuid
        primary_key: true
        default: gen_random_uuid()

  - name: models
    columns:
      - name: id
        type: uuid
        primary_key: true
        default: gen_random_uuid()
      - name: project_id
        type: uuid
        foreign_key:
          references: projects
          column: id
      - name: state
        type: varchar(255)
```

## Utilisation

1. **Configurer la base de données** : Modifiez `config.yaml` selon vos paramètres PostgreSQL

2. **Lancer le programme** :

   ```bash
   go run main.go
   ```

3. **Le programme va** :
   - Charger la configuration depuis `config.yaml`
   - Créer les tables définies
   - Insérer les données d'exemple
   - Exécuter le test de concurrence

## Test de concurrence

Le programme teste deux transactions simultanées :

1. **Transaction T1** : Met à jour la table `models` et reste ouverte 30 secondes
2. **Transaction T2** : Met à jour la table `files` simultanément

Ce test permet de vérifier si PostgreSQL bloque les transactions sur des tables différentes.

## Dépendances

- `github.com/lib/pq` : Driver PostgreSQL
- `github.com/uptrace/bun` : ORM Go
- `gopkg.in/yaml.v3` : Parser YAML

## Installation des dépendances

```bash
go mod tidy
```
