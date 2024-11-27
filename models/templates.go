package models

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates map[string]*template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates[name].Execute(w, data)
}

func NewTemplate() *Templates {
	tmpl := make(map[string]*template.Template)
	tmpl["home"] = template.Must(template.ParseFiles(
		"views/layouts/base.html",
		"views/pages/home.html"))

	tmpl["search"] = template.Must(template.ParseFiles(
		"views/layouts/base.html",
		"views/pages/search.html",
		"views/components/search/search-bar.html"))

	tmpl["results"] = template.Must(template.ParseFiles(
		"views/layouts/base.html",
		"views/pages/results.html",
		"views/components/search/search-bar.html",
		"views/components/results/filters-sidebar.html",
		"views/components/results/results-list.html",
		"views/components/results/result-card.html"))

	return &Templates{
		templates: tmpl,
	}
}
