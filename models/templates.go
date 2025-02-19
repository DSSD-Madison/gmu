package models

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates map[string]*template.Template
}

func splitAtColon(input string) (string, string, bool) {
	// Check if there's a colon in the string
	if strings.Contains(input, ":") {
		// Split the string at the first colon
		parts := strings.SplitN(input, ":", 2)
		return parts[0], parts[1], true
	}
	// Return false if there's no colon
	return input, "", false
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	group, part, usingSubTemplate := splitAtColon(name)
	tmpl, ok := t.templates[group]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	// usingSubTemplate is when we want to use a specific block within a template
	// For example the sidebar block within the results-list
	// We would pass in a name like results:sidebar
	if usingSubTemplate {
		err := tmpl.ExecuteTemplate(w, part, data)
		if err != nil {
			return fmt.Errorf("failed to render sub template %s: %w", name, err)
		}
	} else {
		err := tmpl.Execute(w, data)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", name, err)
		}
	}

	return nil
}

func NewTemplate() *Templates {
	tmpl := make(map[string]*template.Template)
	tmpl["index"] = template.Must(template.ParseFiles(
		"views/index.html",
		"views/search-home.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
	))
	tmpl["search"] = template.Must(template.ParseFiles(
		"views/search.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
		"views/components/skeleton.html",
	))
	tmpl["results"] = template.Must(template.ParseFiles(
		"views/results.html",
		"views/sidecolumn.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
	))
	tmpl["suggestions"] = template.Must(template.ParseFiles(
		"views/suggestions.html",
	))

	return &Templates{
		templates: tmpl,
	}
}
