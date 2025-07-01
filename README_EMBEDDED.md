# Fichiers de Traduction Embarqués

Ce projet utilise le système de fichiers embarqué (embed) de Go pour inclure les fichiers de traduction directement dans le binaire, évitant ainsi les problèmes de fichiers manquants lors de l'installation.

## Architecture

### Structure des fichiers

```
lock-analyser/
├── locales/
│   ├── en.json          # Traductions anglaises
│   ├── fr.json          # Traductions françaises
│   └── embedded.go      # Système de fichiers embarqué
├── i18n/
│   └── translator.go    # Gestionnaire de traductions
└── cmd/
    └── lockanalyzer/
        └── main.go      # Point d'entrée CLI
```

### Fonctionnement

1. **Package `locales`** : Contient le système de fichiers embarqué

   - `embedded.go` : Utilise `//go:embed *.json` pour embarqué tous les fichiers JSON
   - Fournit des fonctions pour accéder aux fichiers embarqués

2. **Package `i18n`** : Gestionnaire de traductions

   - Utilise le package `locales` pour charger les traductions
   - Charge automatiquement tous les fichiers JSON disponibles
   - Gère la détection de langue et la localisation

3. **CLI** : Point d'entrée de l'application
   - Initialise automatiquement le système de traductions
   - Affiche les messages dans la langue appropriée

## Avantages

### ✅ Avantages des fichiers embarqués

1. **Portabilité** : Le binaire contient toutes les traductions
2. **Simplicité d'installation** : Aucun fichier externe requis
3. **Cohérence** : Les traductions sont toujours disponibles
4. **Performance** : Chargement rapide depuis la mémoire
5. **Sécurité** : Pas de manipulation externe des fichiers de traduction

### ❌ Inconvénients de l'ancienne approche

1. **Fichiers manquants** : Risque de fichiers de traduction absents
2. **Chemins complexes** : Recherche dans plusieurs répertoires
3. **Installation complexe** : Nécessité de copier les fichiers de traduction
4. **Incohérence** : Possibilité de versions différentes des traductions

## Utilisation

### Compilation

```bash
# Compilation normale - les fichiers sont automatiquement embarqués
go build -o lockanalyzer cmd/lockanalyzer/main.go
```

### Exécution

```bash
# L'outil fonctionne immédiatement sans fichiers externes
./lockanalyzer -help

# Changement de langue
./lockanalyzer -help -lang=en
./lockanalyzer -help -lang=fr
```

### Tests

```bash
# Tests du système embarqué
go test ./locales/... -v

# Tests de l'outil CLI
go test ./cmd/lockanalyzer/... -v

# Tous les tests
make test
```

## Ajout de nouvelles langues

1. **Créer le fichier de traduction** :

   ```bash
   # Créer locales/es.json pour l'espagnol
   cp locales/fr.json locales/es.json
   # Modifier les traductions dans locales/es.json
   ```

2. **Mettre à jour les tests** :

   ```go
   // Dans les tests, ajouter la nouvelle langue
   expectedFiles := map[string]bool{
       "en.json": false,
       "fr.json": false,
       "es.json": false,  // Nouvelle langue
   }
   ```

3. **Recompiler** :
   ```bash
   go build -o lockanalyzer cmd/lockanalyzer/main.go
   ```

## Détails techniques

### Directive `//go:embed`

```go
//go:embed *.json
var localesFS embed.FS
```

- Embauche tous les fichiers `.json` du répertoire `locales`
- Disponible depuis Go 1.16+
- Les fichiers sont inclus dans le binaire au moment de la compilation

### Chargement des traductions

```go
// Dans i18n/translator.go
localesFS := locales.GetLocalesFS()
entries, err := fs.ReadDir(localesFS, ".")
for _, entry := range entries {
    if strings.HasSuffix(entry.Name(), ".json") {
        content, err := fs.ReadFile(localesFS, entry.Name())
        bundle.MustParseMessageFileBytes(content, entry.Name())
    }
}
```

### Gestion des erreurs

- Si les fichiers embarqués ne sont pas disponibles, l'application affiche un avertissement
- Les clés de traduction manquantes sont remplacées par leur identifiant
- L'application continue de fonctionner même en cas de problème avec les traductions

## Migration depuis l'ancienne approche

### Avant (fichiers externes)

```go
// Recherche dans plusieurs répertoires
possiblePaths := []string{
    "locales",
    "../locales",
    "../../locales",
}
```

### Après (fichiers embarqués)

```go
// Accès direct aux fichiers embarqués
localesFS := locales.GetLocalesFS()
content, err := fs.ReadFile(localesFS, "fr.json")
```

## Maintenance

### Vérification des fichiers embarqués

```bash
# Lister les fichiers embarqués
go run -c 'package main; import "lock-analyser/locales"; func main() { files, _ := locales.ListLocaleFiles(); for _, f := range files { println(f) } }'

# Vérifier le contenu d'un fichier
go run -c 'package main; import "lock-analyser/locales"; func main() { content, _ := locales.GetLocaleFile("fr.json"); println(string(content)) }'
```

### Mise à jour des traductions

1. Modifier les fichiers JSON dans `locales/`
2. Recompiler l'application
3. Les nouvelles traductions sont automatiquement incluses

## Conclusion

L'utilisation du système de fichiers embarqué de Go simplifie considérablement le déploiement et l'utilisation de l'application. Les utilisateurs n'ont plus besoin de se soucier des fichiers de traduction manquants, et l'installation devient plus simple et plus fiable.
