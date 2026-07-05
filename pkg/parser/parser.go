// Package parser provides HTML parsing with a convenient DOM tree API.
package parser

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Node represents an HTML element with simplified access.
type Node struct {
	HTMLNode *html.Node
	Tag      string
	Atom     atom.Atom
	Attrs    map[string]string
	Text     string // text content (for text nodes)
	Type     NodeType
	Children []*Node
	Parent   *Node
}

// NodeType identifies the kind of DOM node.
type NodeType int

const (
	NodeElement NodeType = iota
	NodeText
	NodeDocument
	NodeComment
	NodeDoctype
)

// Document represents a parsed HTML document.
type Document struct {
	Root     *Node
	Title    string
	Meta     map[string]string
	Links    []Link
	Images   []Image
	Scripts  []Script
	Styles   []Style
	Forms    []Form
	Tables   []Table
	Headings []Heading
	Text     string // full text content
}

// Link represents an <a> or <link> element.
type Link struct {
	HREF  string
	Text  string
	Rel   string
	Type  string // "a" or "link"
	Title string
	Attrs map[string]string
}

// Image represents an <img> element.
type Image struct {
	SRC    string
	Alt    string
	Width  string
	Height string
	Title  string
	Attrs  map[string]string
}

// Script represents a <script> element.
type Script struct {
	SRC     string
	Type    string
	Async   bool
	Defer   bool
	Content string
}

// Style represents a <style> element or <link rel="stylesheet">.
type Style struct {
	HREF    string
	Type    string
	Content string
	Media   string
}

// Form represents a <form> element.
type Form struct {
	Action string
	Method string
	Fields []FormField
	ID     string
}

// FormField represents an input/select/textarea in a form.
type FormField struct {
	Type     string // input type or tag name
	Name     string
	Value    string
	Label    string
	Required bool
	ID       string
}

// Table represents a <table> element.
type Table struct {
	Headers []string
	Rows    [][]string
	Caption string
}

// Heading represents a heading element (h1-h6).
type Heading struct {
	Level int
	Text  string
	Node  *Node
}

// Parse parses HTML from a reader and returns a Document.
func Parse(r io.Reader) (*Document, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	result := &Document{
		Meta: make(map[string]string),
	}

	root := convertNode(doc, nil)
	result.Root = root
	result.buildDocument()
	return result, nil
}

// ParseString parses an HTML string and returns a Document.
func ParseString(s string) (*Document, error) {
	return Parse(strings.NewReader(s))
}

// convertNode converts an html.Node to our Node type.
func convertNode(n *html.Node, parent *Node) *Node {
	if n == nil {
		return nil
	}

	node := &Node{
		HTMLNode: n,
		Parent:   parent,
		Attrs:    make(map[string]string),
	}

	switch n.Type {
	case html.ElementNode:
		node.Type = NodeElement
		node.Tag = n.Data
		node.Atom = n.DataAtom
		for _, attr := range n.Attr {
			node.Attrs[attr.Key] = attr.Val
		}
	case html.TextNode:
		node.Type = NodeText
		// Preserve original text including whitespace
		node.Text = n.Data
	case html.DocumentNode:
		node.Type = NodeDocument
	case html.CommentNode:
		node.Type = NodeComment
		node.Text = n.Data
	case html.DoctypeNode:
		node.Type = NodeDoctype
	}

	// Convert children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		child := convertNode(c, node)
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	return node
}

// buildDocument extracts structured information from the DOM tree.
func (d *Document) buildDocument() {
	if d.Root == nil {
		return
	}
	WalkNodes(d.Root, func(n *Node) {
		switch {
		case n.Tag == "title" && n.Parent != nil && n.Parent.Tag == "head":
			d.Title = d.Title + n.TextContent()

		case n.Tag == "meta":
			if name, ok := n.Attrs["name"]; ok {
				d.Meta[name] = n.Attrs["content"]
			}
			if prop, ok := n.Attrs["property"]; ok {
				d.Meta[prop] = n.Attrs["content"]
			}
			if httpEquiv, ok := n.Attrs["http-equiv"]; ok {
				d.Meta[httpEquiv] = n.Attrs["content"]
			}

		case n.Tag == "a":
			d.Links = append(d.Links, Link{
				HREF:  n.Attrs["href"],
				Text:  n.TextContent(),
				Rel:   n.Attrs["rel"],
				Type:  "a",
				Title: n.Attrs["title"],
			})

		case n.Tag == "link":
			d.Links = append(d.Links, Link{
				HREF:  n.Attrs["href"],
				Rel:   n.Attrs["rel"],
				Type:  "link",
				Attrs: n.Attrs,
			})

		case n.Tag == "img":
			d.Images = append(d.Images, Image{
				SRC:    n.Attrs["src"],
				Alt:    n.Attrs["alt"],
				Width:  n.Attrs["width"],
				Height: n.Attrs["height"],
				Title:  n.Attrs["title"],
			})

		case n.Tag == "script":
			d.Scripts = append(d.Scripts, Script{
				SRC:   n.Attrs["src"],
				Type:  n.Attrs["type"],
				Async: n.Attrs["async"] != "",
				Defer: n.Attrs["defer"] != "",
			})

		case n.Tag == "style":
			d.Styles = append(d.Styles, Style{
				Type:  n.Attrs["type"],
				Media: n.Attrs["media"],
			})

		case n.Tag == "link" && n.Attrs["rel"] == "stylesheet":
			d.Styles = append(d.Styles, Style{
				HREF:  n.Attrs["href"],
				Type:  n.Attrs["type"],
				Media: n.Attrs["media"],
			})

		case n.Tag == "form":
			d.Forms = append(d.Forms, Form{
				Action: n.Attrs["action"],
				Method: n.Attrs["method"],
				ID:     n.Attrs["id"],
			})

		case n.Tag == "table":
			d.Tables = append(d.Tables, d.parseTable(n))

		case n.Atom == atom.H1 || n.Atom == atom.H2 || n.Atom == atom.H3 ||
			n.Atom == atom.H4 || n.Atom == atom.H5 || n.Atom == atom.H6:
			level := int(n.Atom - atom.H1 + 1)
			d.Headings = append(d.Headings, Heading{
				Level: level,
				Text:  n.TextContent(),
				Node:  n,
			})
		}
	})

	d.Text = d.Root.TextContent()
}

// parseTable extracts table data from a <table> element.
func (d *Document) parseTable(tableNode *Node) Table {
	t := Table{}

	WalkNodes(tableNode, func(n *Node) {
		switch n.Tag {
		case "caption":
			t.Caption = n.TextContent()
		case "th":
			t.Headers = append(t.Headers, n.TextContent())
		}
	})

	// Parse rows
	WalkNodes(tableNode, func(n *Node) {
		if n.Tag == "tr" {
			var row []string
			for _, child := range n.Children {
				if child.Tag == "td" {
					row = append(row, child.TextContent())
				}
			}
			if len(row) > 0 {
				t.Rows = append(t.Rows, row)
			}
		}
	})

	return t
}

// WalkNodes traverses the DOM tree depth-first, calling fn for each node.
func WalkNodes(n *Node, fn func(*Node)) {
	if n == nil {
		return
	}
	fn(n)
	for _, child := range n.Children {
		WalkNodes(child, fn)
	}
}

// TextContent returns the concatenated text content of a node and its descendants.
func (n *Node) TextContent() string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	walkText(n, &sb)
	return strings.TrimSpace(sb.String())
}

func walkText(n *Node, sb *strings.Builder) {
	if n == nil {
		return
	}
	if n.Type == NodeText && n.Text != "" {
		text := n.Text
		// Trim trailing whitespace to avoid double spaces
		text = strings.TrimRight(text, " 	\n\r")
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}
	for _, child := range n.Children {
		walkText(child, sb)
	}
}

// FindByID finds a node by its id attribute.
func (d *Document) FindByID(id string) *Node {
	var result *Node
	WalkNodes(d.Root, func(n *Node) {
		if n.Attrs["id"] == id {
			result = n
		}
	})
	return result
}

// FindByTag finds all nodes with the given tag name.
func (d *Document) FindByTag(tag string) []*Node {
	var result []*Node
	WalkNodes(d.Root, func(n *Node) {
		if n.Tag == tag {
			result = append(result, n)
		}
	})
	return result
}

// FindByClass finds all nodes with the given class.
func (d *Document) FindByClass(class string) []*Node {
	var result []*Node
	WalkNodes(d.Root, func(n *Node) {
		classes := strings.Fields(n.Attrs["class"])
		for _, c := range classes {
			if c == class {
				result = append(result, n)
				return
			}
		}
	})
	return result
}

// FindBySelector finds nodes matching a simple CSS selector (tag, #id, .class).
func (d *Document) FindBySelector(selector string) []*Node {
	selector = strings.TrimSpace(selector)
	if strings.HasPrefix(selector, "#") {
		id := selector[1:]
		node := d.FindByID(id)
		if node != nil {
			return []*Node{node}
		}
		return nil
	}
	if strings.HasPrefix(selector, ".") {
		class := selector[1:]
		return d.FindByClass(class)
	}
	return d.FindByTag(selector)
}

// OuterHTML returns the HTML representation of a node.
func (n *Node) OuterHTML() string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	renderNode(n, &sb, 0)
	return sb.String()
}

func renderNode(n *Node, sb *strings.Builder, depth int) {
	if n == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	switch n.Type {
	case NodeText:
		if n.Text != "" {
			sb.WriteString(indent)
			sb.WriteString(n.Text)
			sb.WriteString("\n")
		}
	case NodeComment:
		sb.WriteString(indent)
		sb.WriteString("<!-- -->\n")
	case NodeElement:
		sb.WriteString(indent)
		sb.WriteString("<")
		sb.WriteString(n.Tag)
		for k, v := range n.Attrs {
			if v == "" {
				sb.WriteString(" ")
				sb.WriteString(k)
			} else {
				sb.WriteString(" ")
				sb.WriteString(k)
				sb.WriteString("=\"")
				sb.WriteString(v)
				sb.WriteString("\"")
			}
		}
		sb.WriteString(">")

		// Self-closing tags
		selfClosing := map[string]bool{
			"img": true, "br": true, "hr": true, "input": true,
			"meta": true, "link": true, "area": true, "base": true,
			"col": true, "embed": true, "source": true, "track": true, "wbr": true,
		}
		if selfClosing[n.Tag] {
			sb.WriteString("\n")
			return
		}

		if len(n.Children) > 0 {
			sb.WriteString("\n")
			for _, child := range n.Children {
				renderNode(child, sb, depth+1)
			}
			sb.WriteString(indent)
		}
		sb.WriteString("</")
		sb.WriteString(n.Tag)
		sb.WriteString(">\n")
	}
}
