package schematic

import (
	"reflect"
	"strings"
	"testing"
)

var linkTests = []struct {
	Link *Link
	Type string
}{
	{
		Link: &Link{
			Schema: &Schema{
				Properties: map[string]*Schema{
					"string": {
						Type: "string",
					},
				},
				Type:     "object",
				Required: []string{"string"},
			},
		},
		Type: "String string",
	},
	{
		Link: &Link{
			Schema: &Schema{
				Properties: map[string]*Schema{
					"int": {
						Type: "integer",
					},
				},
				Type: "object",
			},
		},
		Type: "Int *int",
	},
}

func TestLinkType(t *testing.T) {
	for i, lt := range linkTests {
		kind := lt.Link.GoType()
		if !strings.Contains(kind, lt.Type) {
			t.Errorf("%d: wants %v, got %v", i, lt.Type, kind)
		}
	}
}

var paramsTests = []struct {
	Schema     *Schema
	Link       *Link
	Order      []string
	Parameters map[string]string
}{
	{
		Schema: &Schema{},
		Link: &Link{
			HRef: NewHRef("/destroy/"),
			Rel:  "destroy",
		},
		Parameters: map[string]string{},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			HRef: NewHRef("/instances/"),
			Rel:  "instances",
		},
		Order:      []string{"lr"},
		Parameters: map[string]string{"lr": "*ListRange"},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			Rel:  "update",
			HRef: NewHRef("/update/"),
			Schema: &Schema{
				Type: "string",
			},
		},
		Order:      []string{"o"},
		Parameters: map[string]string{"o": "string"},
	},
	{
		Schema: &Schema{
			Definitions: map[string]*Schema{
				"struct": {
					Definitions: map[string]*Schema{
						"uuid": {
							Type: "string",
						},
					},
				},
			},
		},
		Link: &Link{
			HRef: NewHRef("/results/{(%23%2Fdefinitions%2Fstruct%2Fdefinitions%2Fuuid)}"),
		},
		Order:      []string{"structUUID"},
		Parameters: map[string]string{"structUUID": "string"},
	},
}

func TestParameters(t *testing.T) {
	for i, pt := range paramsTests {
		pt.Link.Resolve(pt.Schema)
		order, params := pt.Link.Parameters()
		if !reflect.DeepEqual(order, pt.Order) {
			t.Errorf("%d: wants %v, got %v", i, pt.Order, order)
		}
		if !reflect.DeepEqual(params, pt.Parameters) {
			t.Errorf("%d: wants %v, got %v", i, pt.Parameters, params)
		}

	}
}
