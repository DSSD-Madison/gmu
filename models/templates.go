package models

import (
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates map[string]*template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// DEBUG: REMOVE
	log.Println("Rendering template:", name)
	return t.templates[name].Execute(w, data)
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
