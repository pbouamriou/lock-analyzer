package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/pbouamriou/lock-analyzer/locales"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator manages translations for reports
type Translator struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
}

// detectSystemLanguage attempts to detect system language from environment variables
func detectSystemLanguage() string {
	candidates := []string{
		os.Getenv("LANG"),
		os.Getenv("LC_ALL"),
		os.Getenv("LC_MESSAGES"),
	}
	for _, val := range candidates {
		if val == "" {
			continue
		}
		lang := strings.ToLower(val)
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
	}
	return "fr" // fallback
}

// NewTranslator creates a new translator for the specified language
func NewTranslator(lang string) *Translator {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Determine language to use
	chosenLang := lang
	if chosenLang == "" || !IsValidLanguage(chosenLang) {
		chosenLang = detectSystemLanguage()
	}
	if !IsValidLanguage(chosenLang) {
		chosenLang = "fr"
	}

	// Load translation files from embedded filesystem
	localesFS := locales.GetLocalesFS()
	if localesFS != nil {
		// Load all .json files from the embedded filesystem
		entries, err := fs.ReadDir(localesFS, ".")
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					content, err := fs.ReadFile(localesFS, entry.Name())
					if err == nil {
						bundle.MustParseMessageFileBytes(content, entry.Name())
					} else {
						fmt.Printf("Error reading embedded file %s: %v\n", entry.Name(), err)
					}
				}
			}
		} else {
			fmt.Printf("Error reading embedded locales directory: %v\n", err)
		}
	} else {
		fmt.Printf("Warning: No embedded locales filesystem available\n")
	}

	localizer := i18n.NewLocalizer(bundle, chosenLang)

	return &Translator{
		bundle:    bundle,
		localizer: localizer,
	}
}

// T translates a key with optional arguments
func (t *Translator) T(key string, args ...interface{}) string {
	if len(args) == 0 {
		translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
		if err != nil {
			return key
		}
		return translation
	}

	// If arguments are provided, treat them as template data
	templateData := make(map[string]interface{})
	for i, arg := range args {
		templateData[fmt.Sprintf("arg%d", i+1)] = arg
	}

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})
	if err != nil {
		return key
	}
	return translation
}

// TWithData translates a key with specific template data
func (t *Translator) TWithData(key string, data map[string]interface{}) string {
	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return translation
}

// GetAvailableLanguages returns the list of available languages
func GetAvailableLanguages() []string {
	return []string{"fr", "en", "es", "de"}
}

// IsValidLanguage checks if a language is valid
func IsValidLanguage(lang string) bool {
	validLangs := GetAvailableLanguages()
	for _, validLang := range validLangs {
		if lang == validLang {
			return true
		}
	}
	return false
}
