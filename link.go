package schematic

import "fmt"

// Link represents a Link description.
type Link struct {
	Title        string  `json:"title,omitempty"`
	Description  string  `json:"description,omitempty"`
	HRef         *HRef   `json:"href,omitempty"`
	Rel          string  `json:"rel,omitempty"`
	Method       string  `json:"method,omitempty"`
	Schema       *Schema `json:"schema,omitempty"`
	TargetSchema *Schema `json:"targetSchema,omitempty"`
}

// Parameters returns function parameters names and types.
func (l *Link) Parameters() ([]string, map[string]string) {
	if l.HRef == nil {
		// No HRef property
		panic(fmt.Errorf("no href property declared for %s", l.Title))
	}
	var order []string
	params := make(map[string]string)
	for _, name := range l.HRef.Order {
		def := l.HRef.Schemas[name]
		order = append(order, name)
		params[name] = def.GoType()
	}
	switch l.Rel {
	case "update", "create":
		order = append(order, "o")
		params["o"] = l.GoType()
	case "instances":
		order = append(order, "lr")
		params["lr"] = "*ListRange"
	}
	return order, params
}

// Resolve resolve link schema and href.
func (l *Link) Resolve(r *Schema) {
	// if l.Schema != nil {
	// 	l.Schema = l.Schema.Resolve(r)
	// }
	// if l.TargetSchema != nil {
	// 	l.TargetSchema = l.TargetSchema.Resolve(r)
	// }
	l.HRef.Resolve(r)
}

// GoType returns Go type for the given schema as string.
func (l *Link) GoType() string {
	return l.Schema.goType(true, false)
}
