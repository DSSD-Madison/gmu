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

// Returns the splits of the input string split at the first colon character, along with a boolean true to indicate success
//
// Returns input string and an empty string if no colon is found, and a boolean false to indicate failure
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
	tmpls := &Templates{
		templates: tmpl,
	}

	tmpls.registerResponse("index", []string{
		"views/index.html",
		"views/search-home.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
	})
	tmpls.registerResponse("search", []string{
		"views/search.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
		"views/components/skeleton.html",
	})
	tmpls.registerResponse("results", []string{
		"views/results.html",
		"views/sidecolumn.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
	})
	tmpls.registerResponse("suggestions", []string{
		"views/suggestions.html",
	})

	tmpls.registerPage("document", "views/document/document.html")
	tmpls.registerPage("document-edit", "views/document/document-edit.html")
	tmpls.registerPage("document-new", "views/document/document-new.html")
	tmpls.registerPage("document-delete", "views/document/document-delete.html")

	tmpls.registerPartial("document-edit/patch", "edit", "views/document/document-edit.html")
	tmpls.registerPartial("document-new/put", "new", "views/document/document-new.html")
	tmpls.registerPartial("document-delete/delete", "delete", "views/document/document-delete.html")

	return tmpls
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
