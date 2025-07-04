package formatters

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/pbouamriou/lock-analyzer/i18n"
	"github.com/pbouamriou/lock-analyzer/lockanalyzer"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateData holds data for template rendering
type TemplateData struct {
	Data       *lockanalyzer.ReportData
	Translator *i18n.Translator
	Timestamp  time.Time
}

// TemplateFormatter is a base formatter using Go templates
type TemplateFormatter struct {
	translator *i18n.Translator
	template   *template.Template
}

// NewTemplateFormatter creates a new template-based formatter
func NewTemplateFormatter(lang string, templateName string) (*TemplateFormatter, error) {
	// Load template from embedded files
	templateContent, err := templateFS.ReadFile("templates/" + templateName + ".tmpl")
	if err != nil {
		return nil, err
	}

	t, err := template.New("formatter").Funcs(templateFuncs).Parse(string(templateContent))
	if err != nil {
		return nil, err
	}

	return &TemplateFormatter{
		translator: i18n.NewTranslator(lang),
		template:   t,
	}, nil
}

// Format renders the template with the provided data
func (f *TemplateFormatter) Format(data *lockanalyzer.ReportData, output interface{}) error {
	templateData := TemplateData{
		Data:       data,
		Translator: f.translator,
		Timestamp:  time.Now(),
	}

	var buf bytes.Buffer
	if err := f.template.Execute(&buf, templateData); err != nil {
		return err
	}

	// Convert to string and write to output
	if writer, ok := output.(interface{ Write([]byte) (int, error) }); ok {
		_, err := writer.Write(buf.Bytes())
		return err
	}

	return nil
}

// Template functions
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"repeat": func(s string, count int) string {
		result := ""
		for i := 0; i < count; i++ {
			result += s
		}
		return result
	},
	"dict": func(values ...interface{}) map[string]interface{} {
		if len(values)%2 != 0 {
			return nil
		}
		dict := make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				continue
			}
			dict[key] = values[i+1]
		}
		return dict
	},
}
