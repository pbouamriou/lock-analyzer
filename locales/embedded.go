package locales

import (
	"embed"
	"io/fs"
)

//go:embed *.json
var localesFS embed.FS

// GetLocalesFS returns the embedded filesystem containing locale files
func GetLocalesFS() fs.FS {
	return localesFS
}

// GetLocaleFile reads a specific locale file from the embedded filesystem
func GetLocaleFile(filename string) ([]byte, error) {
	return localesFS.ReadFile(filename)
}

// ListLocaleFiles returns all available locale files
func ListLocaleFiles() ([]string, error) {
	entries, err := localesFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
