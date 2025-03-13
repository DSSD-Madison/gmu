package template

import (
	"fmt"
	"io"

	"github.com/labstack/echo/v4"
)

// Templates holds a map of names and Templates to be used for rendering web pages.
// It implements the Echo.Renderer interface to be used for rendering.
type Templates struct {
	templates map[string]*Template
}

// Render executes templates and writes their output to the given writer.
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

// RegisterTemplate registers a Template to the template map.
func (tmpls *Templates) RegisterTemplate(t *Template) {
	fmt.Printf("registered %s\n", t.Name())
	tmpls.templates[t.name] = t
}

// MustRegisterTemplate is a wrapper for registering a template, simply panics on an error.
func (tmpls *Templates) MustRegisterTemplate(t *Template, err error) {
	tmpls.RegisterTemplate(Must(t, err))
}

// RegisterTemplates returns the Echo Renderer with registered templates to be used for rendering web pages.
func RegisterTemplates() *Templates {
	tmpl := make(map[string]*Template)
	tmpls := &Templates{
		templates: tmpl,
	}

	tmpls.RegisterTemplate(
		New("index").
			ParsePage("views/home.html").
			MustParseFiles("views/components/searchbar.html", "views/suggestions.html"),
	)

	tmpls.RegisterTemplate(
		New("search-standalone").
			ParsePage("views/search-home.html").
			MustParseFiles(
				"views/components/searchbar.html",
				"views/suggestions.html",
				"views/search.html",
				"views/components/skeleton.html",
			),
	)

	tmpls.RegisterTemplate(
		New("search").
			ParseResponse(
				"views/search.html",
				"views/components/searchbar.html",
				"views/suggestions.html",
				"views/components/skeleton.html",
			),
	)

	tmpls.RegisterTemplate(
		New("results").
			ParseResponse(
				"views/results.html",
				"views/components/searchbar.html",
				"views/suggestions.html",
				"views/sidecolumn.html",
			),
	)

	tmpls.RegisterTemplate(
		New("suggestions").
			ParseResponse(
				"views/suggestions.html",
			),
	)

	tmpls.RegisterTemplate(
		New("results-container").
			ParsePartial("views/results.html", "results-container"),
	)

	return tmpls
}
