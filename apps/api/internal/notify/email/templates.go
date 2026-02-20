package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*
var templateFS embed.FS

type tmpl struct {
	set *template.Template
}

func loadTemplates() *tmpl {
	set := template.Must(template.New("").ParseFS(templateFS, "templates/*.html"))
	return &tmpl{set: set}
}

func (t *tmpl) render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.set.ExecuteTemplate(&buf, name+".html", data); err != nil {
		return "", fmt.Errorf("rendering template %s: %w", name, err)
	}
	return buf.String(), nil
}
