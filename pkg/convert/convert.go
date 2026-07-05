// Package convert provides HTML conversion to Markdown and plain text.
package convert

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
)

// ToMarkdown converts an HTML document to Markdown format.
func ToMarkdown(doc *parser.Document) string {
	if doc == nil || doc.Root == nil {
		return ""
	}
	var sb strings.Builder
	convertNodeToMarkdown(doc.Root, &sb, 0)
	return strings.TrimSpace(sb.String()) + "\n"
}

func convertNodeToMarkdown(n *parser.Node, sb *strings.Builder, depth int) {
	if n == nil {
		return
	}

	switch n.Type {
	case parser.NodeText:
		if n.Text != "" {
			sb.WriteString(n.Text)
		}

	case parser.NodeComment:
		// Skip comments in markdown

	case parser.NodeElement:
		convertElementToMarkdown(n, sb, depth)

	case parser.NodeDocument:
		for _, child := range n.Children {
			convertNodeToMarkdown(child, sb, depth)
		}
	}
}

func convertElementToMarkdown(n *parser.Node, sb *strings.Builder, depth int) {
	switch n.Tag {
	// Headings
	case "h1":
		sb.WriteString("\n# ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "h2":
		sb.WriteString("\n## ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "h3":
		sb.WriteString("\n### ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "h4":
		sb.WriteString("\n#### ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "h5":
		sb.WriteString("\n##### ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "h6":
		sb.WriteString("\n###### ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")

	// Block elements
	case "p":
		sb.WriteString("\n")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "br":
		sb.WriteString("\n")
	case "hr":
		sb.WriteString("\n---\n\n")
	case "blockquote":
		sb.WriteString("\n> ")
		convertChildren(n, sb, depth)
		sb.WriteString("\n\n")
	case "pre":
		sb.WriteString("\n```\n")
		convertChildren(n, sb, depth)
		sb.WriteString("\n```\n\n")
	case "code":
		if isInsideTag(n, "pre") {
			convertChildren(n, sb, depth)
		} else {
			sb.WriteString("`")
			convertChildren(n, sb, depth)
			sb.WriteString("`")
		}

	// Lists
	case "ul":
		sb.WriteString("\n")
		convertListItems(n, sb, depth, false)
		sb.WriteString("\n")
	case "ol":
		sb.WriteString("\n")
		convertListItems(n, sb, depth, true)
		sb.WriteString("\n")
	case "li":
		// Handled by convertListItems

	// Links and images
	case "a":
		href := n.Attrs["href"]
		title := n.Attrs["title"]
		var linkText strings.Builder
		convertChildren(n, &linkText, depth)
		text := linkText.String()
		if title != "" {
			fmt.Fprintf(sb, "[%s](%s \"%s\")", text, href, title)
		} else {
			fmt.Fprintf(sb, "[%s](%s)", text, href)
		}
	case "img":
		alt := n.Attrs["alt"]
		src := n.Attrs["src"]
		title := n.Attrs["title"]
		if title != "" {
			fmt.Fprintf(sb, "![%s](%s \"%s\")", alt, src, title)
		} else {
			fmt.Fprintf(sb, "![%s](%s)", alt, src)
		}

	// Inline formatting
	case "strong", "b":
		sb.WriteString("**")
		convertChildren(n, sb, depth)
		sb.WriteString("**")
	case "em", "i":
		sb.WriteString("*")
		convertChildren(n, sb, depth)
		sb.WriteString("*")
	case "del", "s", "strike":
		sb.WriteString("~~")
		convertChildren(n, sb, depth)
		sb.WriteString("~~")
	case "u":
		sb.WriteString("<u>")
		convertChildren(n, sb, depth)
		sb.WriteString("</u>")
	case "sup":
		sb.WriteString("<sup>")
		convertChildren(n, sb, depth)
		sb.WriteString("</sup>")
	case "sub":
		sb.WriteString("<sub>")
		convertChildren(n, sb, depth)
		sb.WriteString("</sub>")
	case "mark":
		sb.WriteString("==")
		convertChildren(n, sb, depth)
		sb.WriteString("==")
	case "kbd":
		sb.WriteString("`")
		convertChildren(n, sb, depth)
		sb.WriteString("`")
	case "abbr":
		title := n.Attrs["title"]
		if title != "" {
			var text strings.Builder
			convertChildren(n, &text, depth)
			fmt.Fprintf(sb, "<abbr title=\"%s\">%s</abbr>", title, text.String())
		} else {
			convertChildren(n, sb, depth)
		}

	// Tables
	case "table":
		convertTableToMarkdown(n, sb)

	// Horizontal rule
	case "div", "section", "article", "main", "aside", "nav", "header", "footer":
		convertChildren(n, sb, depth)

	// Skip these
	case "script", "style", "head", "meta", "link", "title":
		// Skip

	default:
		convertChildren(n, sb, depth)
	}
}

func convertChildren(n *parser.Node, sb *strings.Builder, depth int) {
	for _, child := range n.Children {
		convertNodeToMarkdown(child, sb, depth+1)
	}
}

func convertListItems(listNode *parser.Node, sb *strings.Builder, depth int, ordered bool) {
	counter := 1
	for _, child := range listNode.Children {
		if child.Tag == "li" {
			if ordered {
				fmt.Fprintf(sb, "%d. ", counter)
			} else {
				sb.WriteString("- ")
			}
			counter++

			// Process li content
			var content strings.Builder
			for _, c := range child.Children {
				convertNodeToMarkdown(c, &content, depth+1)
			}
			text := strings.TrimSpace(content.String())
			sb.WriteString(text)
			sb.WriteString("\n")
		}
	}
}

func convertTableToMarkdown(tableNode *parser.Node, sb *strings.Builder) {
	sb.WriteString("\n")

	// Find headers (th elements)
	var headers []string
	var rows [][]string

	parser.WalkNodes(tableNode, func(n *parser.Node) {
		if n.Tag == "thead" {
			parser.WalkNodes(n, func(tr *parser.Node) {
				if tr.Tag == "tr" {
					for _, child := range tr.Children {
						if child.Tag == "th" {
							var text strings.Builder
							convertChildren(child, &text, 0)
							headers = append(headers, strings.TrimSpace(text.String()))
						}
					}
				}
			})
		}
	})

	if len(headers) == 0 {
		// Try to find headers from first row
		parser.WalkNodes(tableNode, func(n *parser.Node) {
			if n.Tag == "tr" && len(headers) == 0 {
				for _, child := range n.Children {
					if child.Tag == "th" || child.Tag == "td" {
						var text strings.Builder
						convertChildren(child, &text, 0)
						headers = append(headers, strings.TrimSpace(text.String()))
					}
				}
			}
		})
	}

	// Find rows
	parser.WalkNodes(tableNode, func(n *parser.Node) {
		if n.Tag == "tr" {
			var row []string
			for _, child := range n.Children {
				if child.Tag == "td" {
					var text strings.Builder
					convertChildren(child, &text, 0)
					row = append(row, strings.TrimSpace(text.String()))
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
	})

	// Write header
	if len(headers) > 0 {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(headers, " | "))
		sb.WriteString(" |\n")
		sb.WriteString("| ")
		for range headers {
			sb.WriteString("--- | ")
		}
		sb.WriteString("\n")
	}

	// Write rows
	for _, row := range rows {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(row, " | "))
		sb.WriteString(" |\n")
	}
	sb.WriteString("\n")
}

// isInsideTag checks if a node is inside a specific ancestor tag.
func isInsideTag(n *parser.Node, tag string) bool {
	current := n.Parent
	for current != nil {
		if current.Tag == tag {
			return true
		}
		current = current.Parent
	}
	return false
}

// ToText converts an HTML document to plain text.
func ToText(doc *parser.Document) string {
	if doc == nil || doc.Root == nil {
		return ""
	}
	var sb strings.Builder
	convertNodeToText(doc.Root, &sb)
	result := sb.String()

	// Clean up multiple blank lines
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result) + "\n"
}

func convertNodeToText(n *parser.Node, sb *strings.Builder) {
	if n == nil {
		return
	}

	switch n.Type {
	case parser.NodeText:
		if n.Text != "" {
			sb.WriteString(n.Text)
		}
	case parser.NodeComment:
		// Skip
	case parser.NodeElement:
		convertElementToText(n, sb)
	case parser.NodeDocument:
		for _, child := range n.Children {
			convertNodeToText(child, sb)
		}
	}
}

func convertElementToText(n *parser.Node, sb *strings.Builder) {
	switch n.Tag {
	case "script", "style", "head", "meta", "link", "title":
		return

	// Block elements add line breaks
	case "p", "div", "blockquote", "section", "article",
		"header", "footer", "nav", "aside", "main":
		sb.WriteString("\n")
		for _, child := range n.Children {
			convertNodeToText(child, sb)
		}
		sb.WriteString("\n")

	case "br":
		sb.WriteString("\n")
	case "hr":
		sb.WriteString("\n---\n")

	// Headings add extra line breaks
	case "h1", "h2", "h3", "h4", "h5", "h6":
		sb.WriteString("\n\n")
		for _, child := range n.Children {
			convertNodeToText(child, sb)
		}
		sb.WriteString("\n")

	// Lists
	case "ul", "ol":
		sb.WriteString("\n")
		convertListItemsText(n, sb)
		sb.WriteString("\n")

	case "table":
		convertTableToText(n, sb)

	default:
		for _, child := range n.Children {
			convertNodeToText(child, sb)
		}
	}
}

func convertListItemsText(listNode *parser.Node, sb *strings.Builder) {
	counter := 1
	for _, child := range listNode.Children {
		if child.Tag == "li" {
			if listNode.Tag == "ol" {
				fmt.Fprintf(sb, "%d. ", counter)
			} else {
				sb.WriteString("- ")
			}
			counter++
			for _, c := range child.Children {
				convertNodeToText(c, sb)
			}
			sb.WriteString("\n")
		}
	}
}

func convertTableToText(tableNode *parser.Node, sb *strings.Builder) {
	sb.WriteString("\n")
	var allRows [][]string

	parser.WalkNodes(tableNode, func(n *parser.Node) {
		if n.Tag == "tr" {
			var row []string
			for _, child := range n.Children {
				if child.Tag == "th" || child.Tag == "td" {
					var text strings.Builder
					convertNodeToText(child, &text)
					row = append(row, strings.TrimSpace(text.String()))
				}
			}
			if len(row) > 0 {
				allRows = append(allRows, row)
			}
		}
	})

	if len(allRows) == 0 {
		return
	}

	// Calculate column widths
	maxCols := 0
	for _, row := range allRows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	widths := make([]int, maxCols)
	for _, row := range allRows {
		for i, cell := range row {
			if i < maxCols && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Write rows
	for ri, row := range allRows {
		sb.WriteString("| ")
		for i := 0; i < maxCols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			sb.WriteString(fmt.Sprintf("%-*s", widths[i], cell))
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")

		// Add separator after first row
		if ri == 0 {
			sb.WriteString("| ")
			for _, w := range widths {
				sb.WriteString(strings.Repeat("-", w))
				sb.WriteString(" | ")
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")
}
