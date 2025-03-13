package models

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates map[string]*Template
}

type Template struct {
	name string
	tmpl *template.Template
}

// Parse parses text into the given template
func (t *Template) Parse(text string) (*Template, error) {
	tmpl, err := t.tmpl.Parse(text)
	if err != nil {
		return nil, err
	}
	t.tmpl = tmpl
	return t, nil
}

// Execute executes the template to the given writer
func (t *Template) Execute(w io.Writer, data interface{}) error {
	err := t.tmpl.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

// ExecuteTemplate executes the specified named template to the given io.Writer
func (t *Template) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	err := t.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

// New defines a new template to associate with the given template
func (t *Template) New(name string) *Template {
	t.tmpl = t.tmpl.New(name)
	return t
}

// ParseFiles parses files into the given template.
func (t *Template) ParseFiles(fileNames ...string) (*Template, error) {
	_, err := t.tmpl.ParseFiles(fileNames...)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// New creates a new template with no content associated with it
func New(name string) *Template {
	t := &Template{name: name, tmpl: template.New(name)}

	return t
}

// Name returns the name of the Template
func (t *Template) Name() string {
	return t.tmpl.Name()
}

// Must is a wrapper for template functions. Panics if an error occurs.
func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

// Lookup returns the specified template if it exists, otherwise returns nil
func (t *Template) Lookup(name string) *Template {
	tmpl := t.tmpl.Lookup(name)
	if tmpl != nil {
		return &Template{name: name, tmpl: tmpl}
	}
	return nil
}

// Must  is wrapper function for parsing templates. Panics if it encounters an error
func (t *Template) Must(err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

// MustParseFiles is a wrapper for parsing files. Panics if an error occurs
func (t *Template) MustParseFiles(fileNames ...string) *Template {
	return Must(t.ParseFiles(fileNames...))
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

// RegisterTemplate registers a Template to the template map
func (tmpls *Templates) RegisterTemplate(t *Template) {
	fmt.Printf("registered %s\n", t.Name())
	tmpls.templates[t.name] = t
}

// MustRegisterTemplate is a wrapper for registering a template, simply panics on an error
func (tmpls *Templates) MustRegisterTemplate(t *Template, err error) {
	tmpls.RegisterTemplate(Must(t, err))
}

// WithTitle adds a "title" template to the Template
// Intended for setting the HTML header title, useful for Pages
func (t *Template) WithTitle(title string) *Template {
	t.New(title).Parse(fmt.Sprintf(`{{define "title"}}%s{{end}}`, title))
	return t
}

// ParseResponse parses a collection of files and returns a template
// Intended for larger responses from HTMX, with multiple templates
func (t *Template) ParseResponse(fileNames ...string) *Template {
	_, err := t.ParseFiles(fileNames...)
	if err != nil {
		fmt.Printf("tmpl err %q\n", err)
	}
	return t
}

// ParsePage uses the base html layout to allow for less html boilerplate
// Intended to be used for full-page responses, not "partial" responses like what occur when using HTMX
func (t *Template) ParsePage(page string) *Template {
	name := t.Name()
	fmt.Println(t.Name())
	t, err := New(name).Parse(fmt.Sprintf(`{{block "base" .}}{{end}}`))
	if err != nil {
		fmt.Printf("tmpl err %q\n", err)
	}
	t, err = t.New("base").ParseFiles("views/base.html", page)
	if err != nil {
		fmt.Printf("tmpl err %q\n", err)
	}
	return t
}

// ParsePartial parses a file and returns a single named template from that file
func (t *Template) ParsePartial(partialFile string, partial string) *Template {
	_, err := t.ParseFiles(partialFile)
	if err != nil {
		fmt.Printf("error parsing template: %q", err)
	}
	content := t.Lookup(partial)
	return content
}

func NewTemplate() *Templates {
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
				"views/components/pagination.html",
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
