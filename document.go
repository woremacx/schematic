package schematic

import (
	"bytes"
	"fmt"
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

// Document represents the root Schema.
type Document struct {
	*Schema
}

// Resolve resolves reference inside the schema.
func (d *Document) Resolve(s *Schema) *Schema {
	if len(s.AnyOf) > 0 {
		return d.Resolve(&s.AnyOf[0])
	}
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
			Document   *Document
		}{
			Name:       name,
			Definition: schema,
			Document:   d,
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

// GoType returns the Go type for the given schema as string.
func (d *Document) GoType(s *Schema) string {
	return d.goType(s, true, true)
}

func (d *Document) goType(s *Schema, required bool, force bool) (goType string) {
	// Resolve JSON reference/pointer
	t := d.Resolve(s)
	types := t.Types()
	for _, kind := range types {
		switch kind {
		case "boolean":
			goType = "bool"
		case "string":
			switch t.Format {
			case "date-time":
				goType = "time.Time"
			default:
				goType = "string"
			}
		case "number":
			goType = "float64"
		case "integer":
			goType = "int"
		case "any":
			goType = "interface{}"
		case "array":
			if t.Items != nil {
				goType = "[]" + d.goType(t.Items, required, force)
			} else {
				goType = "[]interface{}"
			}
		case "object":
			// Check if patternProperties exists.
			if t.PatternProperties != nil {
				for _, prop := range t.PatternProperties {
					goType = fmt.Sprintf("map[string]%s", d.GoType(prop))
					break // We don't support more than one pattern for now.
				}
				continue
			}
			buf := bytes.NewBufferString("struct {")
			for _, name := range sortedKeys(t.Properties) {
				prop := d.Resolve(t.Properties[name])
				req := contains(name, t.Required) || force
				templates.ExecuteTemplate(buf, "field.tmpl", struct {
					Definition *Schema
					Name       string
					Required   bool
					Type       string
				}{
					Definition: prop,
					Name:       name,
					Required:   req,
					Type:       d.goType(prop, req, force),
				})
			}
			buf.WriteString("}")
			goType = buf.String()
		case "null":
			continue
		default:
			panic(fmt.Sprintf("unknown type %s", kind))
		}
	}
	if goType == "" {
		panic(fmt.Sprintf("type not found : %s", types))
	}
	// Types allow null
	if contains("null", types) || !(required || force) {
		return "*" + goType
	}
	return goType
}

// LinkType returns Go type for the given link as string.
func (d *Document) LinkType(l *Link) string {
	return d.goType(l.Schema, true, false)
}

// Parameters returns link parameters names and types.
func (d *Document) LinkParameters(l *Link) ([]string, map[string]string) {
	if l.HRef == nil {
		// No HRef property
		panic(fmt.Errorf("no href property declared for %s", l.Title))
	}
	var order []string
	params := make(map[string]string)
	l.HRef.Resolve(d.Schema)
	for _, name := range l.HRef.Order {
		def := d.Resolve(l.HRef.Schemas[name])
		order = append(order, name)
		params[name] = d.GoType(def)
	}
	if l.Schema != nil {
		order = append(order, "o")
		params["o"] = d.LinkType(l)
	}
	if l.TargetSchema != nil {
		targetType := d.Resolve(l.TargetSchema).Types()
		if contains("array", targetType) {
			order = append(order, "lr")
			params["lr"] = "*ListRange"
		}
	}
	return order, params
}

// Values returns link return values types.
func (d *Document) LinkValues(l *Link) (values []string) {
	if l.TargetSchema != nil && l.TargetSchema.Ref != nil {
		name := l.TargetSchema.Ref.Name()
		// s := d.Resolve(l.TargetSchema)
		values = append(values, initialCap(name))
	}
	// switch l.Rel {
	// case "destroy", "empty":
	// 	values = append(values, "error")
	// case "instances":
	// 	values = append(values, fmt.Sprintf("[]*%s", name), "error")
	// default:
	// 	if s.IsCustomType() {
	// 		values = append(values, fmt.Sprintf("*%s", name), "error")
	// 		// } else {
	// 		// values = append(values, s.GoType(), "error")
	// 	}
	// }
	// Append error
	values = append(values, "error")
	return values
}
