package models

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates map[string]*template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return nil
}

func NewTemplate() *Templates {
	tmpl := make(map[string]*template.Template)
	tmpl["index"] = template.Must(template.ParseFiles(
		"views/index.html",
		"views/search-home.html",
	))
	tmpl["search"] = template.Must(template.ParseFiles(
		"views/search.html",
	))
	tmpl["results"] = template.Must(template.ParseFiles(
		"views/results.html",
		"views/sidecolumn.html",
	))

	return &Templates{
		templates: tmpl,
	}
}
