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
	tmpls := &Templates{
		templates: tmpl,
	}

	tmpls.registerResponse("index", []string{
		"views/index.html",
		"views/home.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
	})
	tmpls.registerResponse("search-standalone", []string{
		"views/index.html",
		"views/search-home.html",
		"views/components/searchbar.html",
		"views/suggestions.html",
		"views/search.html",
		"views/components/skeleton.html",
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

	return tmpls
}

func (tmpls *Templates) registerResponse(key string, files []string) {
	tmpls.templates[key] = template.Must(template.ParseFiles(files...))
}

func (tmpls *Templates) registerPage(key string, partialFile string) {
	t, _ := template.New(key).Parse(fmt.Sprintf(`{{block "base" .}}{{end}}`))
	t, err := t.New("base").ParseFiles("views/base.html", partialFile)
	if err != nil {
		fmt.Printf("tmpl err %q\n", err)
	}
	tmpls.templates[key] = t
}

func (tmpls *Templates) registerPartial(key string, partial string, partialFile string) {
	t, err := template.ParseFiles(partialFile)
	if err != nil {
		fmt.Printf("error parsing template: %q", err)
	}
	content := t.Lookup(partial)
	tmpls.templates[key] = content
}
