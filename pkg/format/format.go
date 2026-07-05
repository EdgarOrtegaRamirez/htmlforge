// Package format provides HTML formatting and pretty-printing.
package format

import (
	"strings"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
)

// Formatter formats HTML with configurable indentation and style.
type Formatter struct {
	Indent      string // indentation string (default: "  ")
	MaxLineLen  int    // max line length before wrapping (0 = no wrap)
	IndentAttrs bool   // put each attribute on its own line
	SelfClose   bool   // use self-closing syntax for void elements
	QuoteStyle  string // quote style: "double", "single", "minimal"
}

// Default returns a Formatter with sensible defaults.
func Default() *Formatter {
	return &Formatter{
		Indent:      "  ",
		MaxLineLen:  0,
		IndentAttrs: false,
		SelfClose:   false,
		QuoteStyle:  "double",
	}
}

// FormatHTML pretty-prints an HTML string.
func FormatHTML(html string) string {
	return Default().Format(html)
}

// Format formats an HTML string with the given configuration.
func (f *Formatter) Format(html string) string {
	doc, err := parser.ParseString(html)
	if err != nil {
		return html
	}

	var sb strings.Builder
	f.formatNode(doc.Root, &sb, 0)
	result := sb.String()

	// Clean up multiple blank lines
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result) + "\n"
}

func (f *Formatter) formatNode(n *parser.Node, sb *strings.Builder, depth int) {
	if n == nil {
		return
	}

	switch n.Type {
	case parser.NodeComment:
		indent := f.getIndent(depth)
		sb.WriteString(indent)
		sb.WriteString("<!-- -->\n")

	case parser.NodeText:
		text := n.Text
		if text == "" {
			return
		}
		// Don't indent text that's inline
		sb.WriteString(text)

	case parser.NodeDoctype:
		sb.WriteString("<!doctype html>\n")

	case parser.NodeElement:
		f.formatElement(n, sb, depth)

	case parser.NodeDocument:
		for _, child := range n.Children {
			f.formatNode(child, sb, depth)
		}
	}
}

func (f *Formatter) formatElement(n *parser.Node, sb *strings.Builder, depth int) {
	indent := f.getIndent(depth)

	// Inline elements (don't add newlines around them)
	inlineElements := map[string]bool{
		"a": true, "abbr": true, "b": true, "bdo": true, "br": true,
		"cite": true, "code": true, "dfn": true, "em": true, "i": true,
		"img": true, "input": true, "kbd": true, "label": true, "map": true,
		"object": true, "output": true, "q": true, "samp": true, "script": true,
		"select": true, "small": true, "span": true, "strong": true, "sub": true,
		"sup": true, "textarea": true, "time": true, "var": true,
	}

	// Void elements (no closing tag)
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}

	isInline := inlineElements[n.Tag]
	isVoid := voidElements[n.Tag]

	// Write opening tag
	sb.WriteString(indent)
	sb.WriteString("<")
	sb.WriteString(n.Tag)

	// Write attributes
	f.writeAttrs(n, sb, depth)

	if f.SelfClose && isVoid {
		sb.WriteString(" />")
	} else {
		sb.WriteString(">")
	}

	// Write children
	if len(n.Children) > 0 {
		if !isInline && !isVoid {
			sb.WriteString("\n")
		}
		for _, child := range n.Children {
			if !isInline && !isVoid {
				f.formatNode(child, sb, depth+1)
			} else {
				f.formatNode(child, sb, depth)
			}
		}
		if !isInline && !isVoid {
			sb.WriteString(indent)
		}
	}

	// Write closing tag
	if !isVoid {
		sb.WriteString("</")
		sb.WriteString(n.Tag)
		sb.WriteString(">")
	}

	if !isInline {
		sb.WriteString("\n")
	}
}

func (f *Formatter) writeAttrs(n *parser.Node, sb *strings.Builder, depth int) {
	if len(n.Attrs) == 0 {
		return
	}

	if f.IndentAttrs && len(n.Attrs) > 1 {
		attrIndent := f.getIndent(depth + 1)
		for k, v := range n.Attrs {
			sb.WriteString("\n")
			sb.WriteString(attrIndent)
			sb.WriteString(k)
			if v != "" {
				sb.WriteString("=")
				sb.WriteString(f.quoteValue(v))
			}
		}
		return
	}

	for k, v := range n.Attrs {
		sb.WriteString(" ")
		sb.WriteString(k)
		if v != "" {
			sb.WriteString("=")
			sb.WriteString(f.quoteValue(v))
		}
	}
}

func (f *Formatter) quoteValue(val string) string {
	switch f.QuoteStyle {
	case "single":
		if !strings.Contains(val, "'") {
			return "'" + val + "'"
		}
		return "\"" + strings.ReplaceAll(val, "\"", "&quot;") + "\""
	case "minimal":
		if !strings.ContainsAny(val, " \"'=<>\t\n") {
			return val
		}
		if !strings.Contains(val, "\"") {
			return "\"" + val + "\""
		}
		return "'" + val + "'"
	default: // "double"
		if !strings.Contains(val, "\"") {
			return "\"" + val + "\""
		}
		return "'" + val + "'"
	}
}

func (f *Formatter) getIndent(depth int) string {
	var sb strings.Builder
	for i := 0; i < depth; i++ {
		sb.WriteString(f.Indent)
	}
	return sb.String()
}

// Minify removes unnecessary whitespace from HTML (without parsing).
func Minify(html string) string {
	var sb strings.Builder
	inTag := false
	inScript := false
	inStyle := false
	prevWasSpace := false

	for i := 0; i < len(html); i++ {
		ch := html[i]

		if ch == '<' && i+1 < len(html) {
			// Check for script/style tags
			remaining := html[i:]
			if strings.HasPrefix(remaining, "<script") {
				inScript = true
			} else if strings.HasPrefix(remaining, "</script") {
				inScript = false
			} else if strings.HasPrefix(remaining, "<style") {
				inStyle = true
			} else if strings.HasPrefix(remaining, "</style") {
				inStyle = false
			}

			inTag = true
			prevWasSpace = false
			sb.WriteByte(ch)
			continue
		}

		if ch == '>' && inTag {
			inTag = false
			prevWasSpace = false
			sb.WriteByte(ch)
			continue
		}

		if inScript || inStyle {
			sb.WriteByte(ch)
			prevWasSpace = false
			continue
		}

		if inTag {
			sb.WriteByte(ch)
			prevWasSpace = false
			continue
		}

		// Outside tags - collapse whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if !prevWasSpace {
				sb.WriteByte(' ')
				prevWasSpace = true
			}
		} else {
			prevWasSpace = false
			sb.WriteByte(ch)
		}
	}

	return strings.TrimSpace(sb.String())
}
