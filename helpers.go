package schematic

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

var helpers = template.FuncMap{
	"initialCap": initialCap,
	"initialLow": initialLow,
	"methodCap":  methodCap,
	"asComment":  asComment,
	"jsonTag":    jsonTag,
	"params":     params,
	"args":       args,
	"values":     values,
	"goType":     goType,
	"linkType":   linkType,
	"targetType": targetType,
}

var (
	newlines  = regexp.MustCompile(`(?m:\s*$)`)
	acronyms  = regexp.MustCompile(`(Url|Http|Id|Io|Uuid|Api|Uri|Ssl|Cname|Oauth|Otp)$`)
	camelcase = regexp.MustCompile(`(?m)[-.$/:_{}\s]`)
)

func goType(d *Document, p *Schema) string {
	return d.GoType(p)
}

func linkType(d *Document, l *Link) string {
	return d.LinkType(l)
}

func targetType(d *Document, l *Link) string {
	return d.TargetType(l)
}

func required(n string, def *Schema) bool {
	return contains(n, def.Required)
}

func jsonTag(n string, required bool) string {
	tags := []string{n}
	if !required {
		tags = append(tags, "omitempty")
	}
	return fmt.Sprintf("`json:\"%s\"`", strings.Join(tags, ","))
}

func contains(n string, r []string) bool {
	for _, r := range r {
		if r == n {
			return true
		}
	}
	return false
}

func initialCap(ident string) string {
	if ident == "" {
		panic("blank identifier")
	}
	return depunct(ident, true)
}

func methodCap(ident string) string {
	return initialCap(strings.ToLower(ident))
}

func initialLow(ident string) string {
	if ident == "" {
		panic("blank identifier")
	}
	return depunct(ident, false)
}

func depunct(ident string, initialCap bool) string {
	matches := camelcase.Split(ident, -1)
	for i, m := range matches {
		if initialCap || i > 0 {
			m = capFirst(m)
		}
		matches[i] = acronyms.ReplaceAllStringFunc(m, func(c string) string {
			if len(c) > 4 {
				return strings.ToUpper(c[:2]) + c[2:]
			}
			return strings.ToUpper(c)
		})
	}
	return strings.Join(matches, "")
}

func capFirst(ident string) string {
	r, n := utf8.DecodeRuneInString(ident)
	return string(unicode.ToUpper(r)) + ident[n:]
}

func asComment(c string) string {
	var buf bytes.Buffer
	const maxLen = 70
	removeNewlines := func(s string) string {
		return strings.Replace(s, "\n", "\n// ", -1)
	}
	for len(c) > 0 {
		line := c
		if len(line) < maxLen {
			fmt.Fprintf(&buf, "// %s\n", removeNewlines(line))
			break
		}
		line = line[:maxLen]
		si := strings.LastIndex(line, " ")
		if si != -1 {
			line = line[:si]
		}
		fmt.Fprintf(&buf, "// %s\n", removeNewlines(line))
		c = c[len(line):]
		if si != -1 {
			c = c[1:]
		}
	}
	return buf.String()
}

func values(d *Document, l *Link) string {
	v := d.LinkValues(l)
	return strings.Join(v, ", ")
}

func params(d *Document, l *Link) string {
	var p []string
	order, params := d.LinkParameters(l)
	for _, n := range order {
		p = append(p, fmt.Sprintf("%s %s", initialLow(n), params[n]))
	}
	return strings.Join(p, ", ")
}

func args(h *HRef) string {
	return strings.Join(h.Order, ", ")
}

func sortedKeys(m map[string]*Schema) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}
