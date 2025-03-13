package template

import (
	"fmt"
	"html/template"
	"io"
)

// Template is a wrapper around template.Template to add new functions.
type Template struct {
	name string
	tmpl *template.Template
}

// Parse parses text into the given template.
func (t *Template) Parse(text string) (*Template, error) {
	tmpl, err := t.tmpl.Parse(text)
	if err != nil {
		return nil, err
	}
	t.tmpl = tmpl
	return t, nil
}

// Execute executes the template to the given writer.
func (t *Template) Execute(w io.Writer, data interface{}) error {
	err := t.tmpl.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

// ExecuteTemplate executes the specified named template to the given io.Writer.
func (t *Template) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	err := t.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

// New defines a new template to associate with the given template.
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

// MustParseFiles is a wrapper for parsing files. Panics if an error occurs.
func (t *Template) MustParseFiles(fileNames ...string) *Template {
	return Must(t.ParseFiles(fileNames...))
}

// WithTitle adds a "title" template to the Template
// Intended for setting the HTML header title, useful for Pages.
func (t *Template) WithTitle(title string) *Template {
	t.New(title).Parse(fmt.Sprintf(`{{define "title"}}%s{{end}}`, title))
	return t
}

// ParseResponse parses a collection of files and returns a template.
// Intended for larger responses from HTMX, with multiple templates.
func (t *Template) ParseResponse(fileNames ...string) *Template {
	_, err := t.ParseFiles(fileNames...)
	if err != nil {
		fmt.Printf("tmpl err %q\n", err)
	}
	return t
}

// ParsePage uses the base html layout to allow for less html boilerplate.
// Intended to be used for full-page responses, not "partial" responses like what occur when using HTMX.
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
