package models

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

type TemplatesNew struct {
	templates map[string]*Template
}

type Templates struct {
	templates map[string]*template.Template
}

type Template struct {
	name string
	tmpl *template.Template
}

func (t *Template) Parse(text string) (*Template, error) {
	tmpl, err := t.tmpl.Parse(text)
	if err != nil {
		return nil, err
	}
	t.tmpl = tmpl
	return t, nil
}

func (t *Template) Execute(w io.Writer, data interface{}) error {
	err := t.tmpl.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

func (t *Template) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	err := t.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

func (t *Template) New(name string) *Template {
	t.tmpl = t.tmpl.New(name)
	return t
}

func (t *Template) ParseFiles(fileNames ...string) (*Template, error) {
	_, err := t.tmpl.ParseFiles(fileNames...)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func New(name string) *Template {
	t := &Template{name: name, tmpl: template.New(name)}

	return t
}

func (t *Template) Name() string {
	return t.tmpl.Name()
}

func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

func (t *Template) Lookup(name string) *Template {
	tmpl := t.tmpl.Lookup(name)
	if tmpl != nil {
		return &Template{name: name, tmpl: tmpl}
	}
	return nil
}

func (t *Template) Must(err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

func (t *Template) MustParseFiles(fileNames ...string) *Template {
	return Must(t.ParseFiles(fileNames...))
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

func (t *TemplatesNew) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
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

func NewTemplate() *TemplatesNew {
	tmpl := make(map[string]*Template)
	tmpls := &TemplatesNew{
		templates: tmpl,
	}
	// templs := &Templates{
	// 	templates: make(map[string]*template.Template),
	// }

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

	defined := tmpls.templates["results"].tmpl.DefinedTemplates()
	fmt.Println(defined)
	fmt.Println(tmpls.templates["results"].Name())
	tmpls.templates["results"].Execute(os.Stdout, nil)

	tmpls.RegisterTemplate(
		New("suggestions").
			ParseResponse(
				"views/suggestions.html",
			),
	)

	for k, v := range tmpls.templates {
		fmt.Println(k, v)
	}

	// tmpls.registerResponse("index", []string{
	// 	"views/index.html",
	// 	"views/home.html",
	// 	"views/components/searchbar.html",
	// 	"views/suggestions.html",
	// })
	// tmpls.registerTitle("index", "Better Evidence Project - Home")
	// tmpls.registerResponse("search-standalone", []string{
	// 	"views/index.html",
	// 	"views/search-home.html",
	// 	"views/components/searchbar.html",
	// 	"views/suggestions.html",
	// 	"views/search.html",
	// 	"views/components/skeleton.html",
	// })
	// tmpls.registerResponse("search", []string{
	// 	"views/search.html",
	// 	"views/components/searchbar.html",
	// 	"views/suggestions.html",
	// 	"views/components/skeleton.html",
	// })
	// tmpls.registerResponse("results", []string{
	// 	"views/results.html",
	// 	"views/sidecolumn.html",
	// 	"views/components/searchbar.html",
	// 	"views/suggestions.html",
	// })
	// tmpls.registerResponse("suggestions", []string{
	// 	"views/suggestions.html",
	// })

	return tmpls
}

func (tmpls *TemplatesNew) RegisterTemplate(t *Template) {
	fmt.Printf("registered %s\n", t.Name())
	tmpls.templates[t.name] = t
}

func (tmpls *TemplatesNew) MustRegisterTemplate(t *Template, err error) {
	tmpls.RegisterTemplate(Must(t, err))
}

func (t *Template) WithTitle(title string) *Template {
	tmpl := template.New("")
	t.New(title).Parse(fmt.Sprintf(`{{define "title"}}%s{{end}}`, title))
	fmt.Printf(tmpl.DefinedTemplates())
	return t
}

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

func (t *Template) ParsePartial(partial string, partialFile string) *Template {
	t, err := t.ParseFiles(partialFile)
	if err != nil {
		fmt.Printf("error parsing template: %q", err)
	}
	content := t.Lookup(partial)
	return content
}

// Registers a template by parsing a list of templates together
func (tmpls *Templates) registerResponse(key string, files []string) {
	tmpls.templates[key] = template.Must(template.ParseFiles(files...))
}

// Registers a page using the base HTML template base.html.
func (tmpls *Templates) registerPage(key string, partialFile string) {
	t, _ := template.New(key).Parse(fmt.Sprintf(`{{block "base" .}}{{end}}`))
	t, err := t.New("base").ParseFiles("views/base.html", partialFile)
	if err != nil {
		fmt.Printf("tmpl err %q\n", err)
	}
	tmpls.templates[key] = t
}

// Registers a key-accessible template associated with a template defined in the given partialFile
func (tmpls *Templates) registerPartial(key string, partial string, partialFile string) {
	t, err := template.ParseFiles(partialFile)
	if err != nil {
		fmt.Printf("error parsing template: %q", err)
	}
	content := t.Lookup(partial)
	tmpls.templates[key] = content
}
