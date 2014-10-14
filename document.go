package schematic

import (
	"bytes"
	"go/format"
	"strings"
	"text/template"

	bundle "github.com/interagent/schematic/templates"
)

var templates *template.Template

func init() {
	templates = template.New("package.tmpl").Funcs(helpers)
	templates = template.Must(bundle.Parse(templates))
}

type Document struct {
	*Schema
}

// Resolve resolves reference inside the schema.
func (d *Document) Resolve(s *Schema) *Schema {
	if s.Ref != nil {
		return s.Ref.Resolve(d.Schema)
	}
	return s
}

// Generate generates code according to the schema.
func (d *Document) Generate() ([]byte, error) {
	var buf bytes.Buffer

	name := strings.ToLower(strings.Split(d.Title, " ")[0])
	templates.ExecuteTemplate(&buf, "package.tmpl", name)

	// TODO: Check if we need time.
	templates.ExecuteTemplate(&buf, "imports.tmpl", []string{
		"encoding/json", "fmt", "io", "reflect",
		"net/http", "runtime", "time", "bytes",
	})
	templates.ExecuteTemplate(&buf, "service.tmpl", struct {
		Name    string
		URL     string
		Version string
	}{
		Name:    name,
		URL:     d.URL(),
		Version: d.Version,
	})

	for _, name := range sortedKeys(d.Properties) {
		schema := d.Resolve(d.Properties[name])
		// Skipping definitions because there is no links, nor properties.
		if schema.Links == nil && schema.Properties == nil {
			continue
		}

		context := struct {
			Name       string
			Definition *Schema
		}{
			Name:       name,
			Definition: schema,
		}

		templates.ExecuteTemplate(&buf, "struct.tmpl", context)
		templates.ExecuteTemplate(&buf, "funcs.tmpl", context)
	}

	// Remove blank lines added by text/template
	bytes := newlines.ReplaceAll(buf.Bytes(), []byte(""))

	// Format sources
	clean, err := format.Source(bytes)
	if err != nil {
		return buf.Bytes(), err
	}
	return clean, nil
}
