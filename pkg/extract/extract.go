// Package extract provides content extraction from HTML documents.
package extract

import (
	"strings"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
)

// Extractor extracts structured content from HTML documents.
type Extractor struct {
	doc *parser.Document
}

// New creates a new Extractor for the given document.
func New(doc *parser.Document) *Extractor {
	return &Extractor{doc: doc}
}

// TextOptions controls text extraction behavior.
type TextOptions struct {
	IncludeScripts bool
	IncludeStyles  bool
	IncludeComments bool
	TrimWhitespace bool
	MinLength      int // minimum text node length to include
}

// DefaultTextOptions returns sensible defaults for text extraction.
func DefaultTextOptions() TextOptions {
	return TextOptions{
		IncludeScripts:  false,
		IncludeStyles:   false,
		IncludeComments: false,
		TrimWhitespace:  true,
		MinLength:       0,
	}
}

// Text extracts all visible text from the document.
func (e *Extractor) Text(opts TextOptions) string {
	var sb strings.Builder
	extractText(e.doc.Root, &opts, &sb)
	result := sb.String()
	if opts.TrimWhitespace {
		// Normalize whitespace
		lines := strings.Split(result, "\n")
		var cleaned []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		result = strings.Join(cleaned, "\n")
	}
	return result
}

func extractText(n *parser.Node, opts *TextOptions, sb *strings.Builder) {
	if n == nil {
		return
	}

	// Skip script and style content
	if !opts.IncludeScripts && n.Tag == "script" {
		return
	}
	if !opts.IncludeStyles && n.Tag == "style" {
		return
	}
	if !opts.IncludeComments && n.Type == parser.NodeComment {
		return
	}

	if n.Type == parser.NodeText && n.Text != "" {
		text := n.Text
		if opts.TrimWhitespace {
			text = strings.TrimSpace(text)
		}
		if len(text) >= opts.MinLength {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}

	// Add line breaks for block elements
	blockElements := map[string]bool{
		"p": true, "div": true, "br": true, "hr": true,
		"h1": true, "h2": true, "h3": true, "h4": true,
		"h5": true, "h6": true, "li": true, "tr": true,
		"blockquote": true, "pre": true, "section": true,
		"article": true, "header": true, "footer": true,
		"nav": true, "aside": true, "main": true,
	}

	if blockElements[n.Tag] {
		sb.WriteString("\n")
	}

	for _, child := range n.Children {
		extractText(child, opts, sb)
	}
}

// Links extracts all links from the document.
func (e *Extractor) Links() []parser.Link {
	return e.doc.Links
}

// Images extracts all images from the document.
func (e *Extractor) Images() []parser.Image {
	return e.doc.Images
}

// Scripts extracts all script references from the document.
func (e *Extractor) Scripts() []parser.Script {
	return e.doc.Scripts
}

// Styles extracts all style references from the document.
func (e *Extractor) Styles() []parser.Style {
	return e.doc.Styles
}

// Forms extracts all forms from the document.
func (e *Extractor) Forms() []parser.Form {
	return e.doc.Forms
}

// Tables extracts all tables from the document.
func (e *Extractor) Tables() []parser.Table {
	return e.doc.Tables
}

// Headings extracts all headings from the document.
func (e *Extractor) Headings() []parser.Heading {
	return e.doc.Headings
}

// HeadingTree returns headings organized in a tree structure.
type HeadingNode struct {
	Level    int
	Text     string
	Children []*HeadingNode
}

// HeadingTree returns headings organized hierarchically.
func (e *Extractor) HeadingTree() *HeadingNode {
	headings := e.doc.Headings
	if len(headings) == 0 {
		return nil
	}

	root := &HeadingNode{Level: 0, Text: "Root"}
	stack := []*HeadingNode{root}

	for _, h := range headings {
		node := &HeadingNode{Level: h.Level, Text: h.Text}

		// Pop stack until we find a parent
		for len(stack) > 1 && stack[len(stack)-1].Level >= h.Level {
			stack = stack[:len(stack)-1]
		}

		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)
		stack = append(stack, node)
	}

	return root
}

// Meta extracts metadata from the document.
func (e *Extractor) Meta() map[string]string {
	return e.doc.Meta
}

// Title returns the document title.
func (e *Extractor) Title() string {
	return e.doc.Title
}

// OpenGraph extracts Open Graph metadata.
func (e *Extractor) OpenGraph() map[string]string {
	result := make(map[string]string)
	for key, val := range e.doc.Meta {
		if strings.HasPrefix(key, "og:") {
			result[key] = val
		}
	}
	return result
}

// TwitterCards extracts Twitter Card metadata.
func (e *Extractor) TwitterCards() map[string]string {
	result := make(map[string]string)
	for key, val := range e.doc.Meta {
		if strings.HasPrefix(key, "twitter:") {
			result[key] = val
		}
	}
	return result
}

// UniqueLinks returns deduplicated links.
func (e *Extractor) UniqueLinks() []parser.Link {
	seen := make(map[string]bool)
	var result []parser.Link
	for _, link := range e.doc.Links {
		if link.HREF != "" && !seen[link.HREF] {
			seen[link.HREF] = true
			result = append(result, link)
		}
	}
	return result
}

// InternalLinks returns links to the same domain.
func (e *Extractor) InternalLinks(domain string) []parser.Link {
	var result []parser.Link
	for _, link := range e.doc.Links {
		href := link.HREF
		if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") ||
			strings.Contains(href, domain) {
			result = append(result, link)
		}
	}
	return result
}

// ExternalLinks returns links to external domains.
func (e *Extractor) ExternalLinks(domain string) []parser.Link {
	var result []parser.Link
	for _, link := range e.doc.Links {
		href := link.HREF
		if strings.HasPrefix(href, "http") && !strings.Contains(href, domain) {
			result = append(result, link)
		}
	}
	return result
}

// ImagesWithoutAlt returns images missing alt text.
func (e *Extractor) ImagesWithoutAlt() []parser.Image {
	var result []parser.Image
	for _, img := range e.doc.Images {
		if img.Alt == "" {
			result = append(result, img)
		}
	}
	return result
}

// BrokenAnchors returns anchor elements without href.
func (e *Extractor) BrokenAnchors() []*parser.Node {
	var result []*parser.Node
	parser.WalkNodes(e.doc.Root, func(n *parser.Node) {
		if n.Tag == "a" && n.Attrs["href"] == "" {
			result = append(result, n)
		}
	})
	return result
}
