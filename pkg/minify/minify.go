// Package minify provides HTML minification.
package minify

import (
	"strings"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
)

// Minifier removes unnecessary whitespace and optional tags from HTML.
type Minifier struct {
	RemoveComments      bool
	RemoveOptionalTags  bool
	CollapseWhitespace  bool
	MinifyInlineCSS     bool
	MinifyInlineJS      bool
	RemoveEmptyAttrs    bool
	SortAttrs           bool
}

// Default returns a Minifier with sensible defaults.
func Default() *Minifier {
	return &Minifier{
		RemoveComments:     true,
		RemoveOptionalTags: false,
		CollapseWhitespace: true,
		MinifyInlineCSS:    false,
		MinifyInlineJS:     false,
		RemoveEmptyAttrs:   true,
		SortAttrs:          false,
	}
}

// MinifyHTML minifies an HTML string.
func MinifyHTML(html string) string {
	return Default().Minify(html)
}

// Minify minifies an HTML string with the given configuration.
func (m *Minifier) Minify(html string) string {
	doc, err := parser.ParseString(html)
	if err != nil {
		return html
	}

	var sb strings.Builder
	minifyNode(doc.Root, m, &sb)
	return sb.String()
}

func minifyNode(n *parser.Node, m *Minifier, sb *strings.Builder) {
	if n == nil {
		return
	}

	switch n.Type {
	case parser.NodeComment:
		if !m.RemoveComments {
			sb.WriteString("<!--")
			sb.WriteString(n.Text)
			sb.WriteString("-->")
		}
		return

	case parser.NodeText:
		text := n.Text
		if text == "" {
			return
		}
		if m.CollapseWhitespace {
			text = collapseWhitespace(text)
		}
		sb.WriteString(text)
		return

	case parser.NodeDoctype:
		sb.WriteString("<!doctype html>")
		return

	case parser.NodeElement:
		minifyElement(n, m, sb)

	case parser.NodeDocument:
		for _, child := range n.Children {
			minifyNode(child, m, sb)
		}
	}
}

func minifyElement(n *parser.Node, m *Minifier, sb *strings.Builder) {
	// Self-closing tags
	selfClosing := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}

	// Optional end tags
	optionalEnd := map[string]bool{
		"li": true, "dt": true, "dd": true, "p": true,
		"tr": true, "td": true, "th": true, "thead": true,
		"tbody": true, "tfoot": true, "colgroup": true,
		"option": true,
	}

	// Skip script/style content
	if n.Tag == "script" || n.Tag == "style" {
		sb.WriteString("<")
		sb.WriteString(n.Tag)
		writeAttrs(n, m, sb)
		sb.WriteString(">")
		for _, child := range n.Children {
			if child.Type == parser.NodeText {
				text := child.Text
				if text == "" {
					continue
				}
				if n.Tag == "style" && m.MinifyInlineCSS {
					text = minifyCSS(text)
				}
				if n.Tag == "script" && m.MinifyInlineJS {
					text = minifyJS(text)
				}
				sb.WriteString(text)
			}
		}
		sb.WriteString("</")
		sb.WriteString(n.Tag)
		sb.WriteString(">")
		return
	}

	// Write opening tag
	sb.WriteString("<")
	sb.WriteString(n.Tag)
	writeAttrs(n, m, sb)

	if selfClosing[n.Tag] {
		sb.WriteString(">")
		return
	}

	sb.WriteString(">")

	// Write children
	for _, child := range n.Children {
		minifyNode(child, m, sb)
	}

	// Write closing tag
	if !(m.RemoveOptionalTags && optionalEnd[n.Tag]) {
		sb.WriteString("</")
		sb.WriteString(n.Tag)
		sb.WriteString(">")
	}
}

func writeAttrs(n *parser.Node, m *Minifier, sb *strings.Builder) {
	attrs := n.Attrs

	// Sort attributes if requested
	if m.SortAttrs && len(attrs) > 1 {
		type kv struct {
			key, val string
		}
		sorted := make([]kv, 0, len(attrs))
		for k, v := range attrs {
			sorted = append(sorted, kv{k, v})
		}
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].key > sorted[j].key {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		for _, a := range sorted {
			if m.RemoveEmptyAttrs && a.val == "" && !isBooleanAttr(a.key) {
				continue
			}
			writeAttr(a.key, a.val, sb)
		}
		return
	}

	for k, v := range attrs {
		if m.RemoveEmptyAttrs && v == "" && !isBooleanAttr(k) {
			continue
		}
		writeAttr(k, v, sb)
	}
}

func writeAttr(key, val string, sb *strings.Builder) {
	if val == "" || isBooleanAttr(key) {
		sb.WriteString(" ")
		sb.WriteString(key)
		return
	}

	sb.WriteString(" ")
	sb.WriteString(key)
	sb.WriteString("=\"")
	// Use minimal quoting
	if !strings.Contains(val, "\"") {
		sb.WriteString(val)
	} else if !strings.Contains(val, "'") {
		sb.WriteString("'")
		sb.WriteString(val)
		sb.WriteString("'")
	} else {
		sb.WriteString(strings.ReplaceAll(val, "\"", "&quot;"))
	}
	sb.WriteString("\"")
}

func isBooleanAttr(attr string) bool {
	boolAttrs := map[string]bool{
		"async": true, "autofocus": true, "autoplay": true,
		"checked": true, "controls": true, "default": true,
		"defer": true, "disabled": true, "download": true,
		"formnovalidate": true, "hidden": true, "ismap": true,
		"itemscope": true, "loop": true, "multiple": true,
		"muted": true, "nomodule": true, "novalidate": true,
		"open": true, "playsinline": true, "readonly": true,
		"required": true, "reversed": true, "selected": true,
	}
	return boolAttrs[attr]
}

func collapseWhitespace(s string) string {
	// Replace multiple whitespace with single space
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func minifyCSS(css string) string {
	// Basic CSS minification
	css = strings.TrimSpace(css)
	// Remove comments
	for strings.Contains(css, "/*") {
		start := strings.Index(css, "/*")
		end := strings.Index(css[start:], "*/")
		if end == -1 {
			break
		}
		css = css[:start] + css[start+end+2:]
	}
	// Remove whitespace around special characters
	replacer := strings.NewReplacer(
		" {", "{",
		"{ ", "{",
		" }", "}",
		"} ", "}",
		" ;", ";",
		"; ", ";",
		" :", ":",
		": ", ":",
		" ,", ",",
		", ", ",",
		"\n", "",
		"\t", "",
		"  ", " ",
	)
	return replacer.Replace(css)
}

func minifyJS(js string) string {
	// Very basic JS minification - just remove comments and excess whitespace
	js = strings.TrimSpace(js)
	// Remove single-line comments (simple approach)
	lines := strings.Split(js, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//") {
			continue
		}
		// Don't remove URLs with //
		if idx := strings.Index(line, " //"); idx > 0 {
			line = line[:idx]
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}
