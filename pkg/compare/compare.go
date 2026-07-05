// Package compare provides semantic HTML comparison.
package compare

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
)

// ChangeType identifies the type of change.
type ChangeType int

const (
	Added ChangeType = iota
	Removed
	Modified
	Unchanged
)

func (c ChangeType) String() string {
	switch c {
	case Added:
		return "added"
	case Removed:
		return "removed"
	case Modified:
		return "modified"
	case Unchanged:
		return "unchanged"
	default:
		return "unknown"
	}
}

// Diff represents a difference between two HTML documents.
type Diff struct {
	Type    ChangeType
	Path    string
	Old     *parser.Node
	New     *parser.Node
	Details string
}

// Compare compares two HTML documents and returns a list of differences.
func Compare(oldDoc, newDoc *parser.Document) []Diff {
	var diffs []Diff
	compareNodes(oldDoc.Root, newDoc.Root, "html", &diffs)
	return diffs
}

func compareNodes(old, new *parser.Node, path string, diffs *[]Diff) {
	if old == nil && new == nil {
		return
	}

	if old == nil {
		*diffs = append(*diffs, Diff{
			Type:    Added,
			Path:    path,
			New:     new,
			Details: fmt.Sprintf("Added <%s> element", new.Tag),
		})
		return
	}

	if new == nil {
		*diffs = append(*diffs, Diff{
			Type:    Removed,
			Path:    path,
			Old:     old,
			Details: fmt.Sprintf("Removed <%s> element", old.Tag),
		})
		return
	}

	// Handle Document nodes - compare children
	if old.Type == parser.NodeDocument && new.Type == parser.NodeDocument {
		compareChildren(old, new, path, diffs)
		return
	}

	// Both exist - compare
	if old.Type != new.Type {
		*diffs = append(*diffs, Diff{
			Type:    Modified,
			Path:    path,
			Old:     old,
			New:     new,
			Details: fmt.Sprintf("Node type changed: %d → %d", old.Type, new.Type),
		})
		return
	}

	// Compare text nodes
	if old.Type == parser.NodeText {
		if old.Text != new.Text {
			*diffs = append(*diffs, Diff{
				Type:    Modified,
				Path:    path,
				Old:     old,
				New:     new,
				Details: fmt.Sprintf("Text changed: %q → %q", old.Text, new.Text),
			})
		}
		return
	}

	// Compare element nodes
	if old.Type == parser.NodeElement {
		if old.Tag != new.Tag {
			*diffs = append(*diffs, Diff{
				Type:    Modified,
				Path:    path,
				Old:     old,
				New:     new,
				Details: fmt.Sprintf("Tag changed: <%s> → <%s>", old.Tag, new.Tag),
			})
			return
		}

		// Compare attributes
		compareAttrs(old, new, path, diffs)

		// Compare children
		compareChildren(old, new, path, diffs)
	}
}

func compareAttrs(old, new *parser.Node, path string, diffs *[]Diff) {
	allKeys := make(map[string]bool)
	for k := range old.Attrs {
		allKeys[k] = true
	}
	for k := range new.Attrs {
		allKeys[k] = true
	}

	for key := range allKeys {
		oldVal, oldExists := old.Attrs[key]
		newVal, newExists := new.Attrs[key]

		if !oldExists {
			*diffs = append(*diffs, Diff{
				Type:    Added,
				Path:    fmt.Sprintf("%s/@%s", path, key),
				New:     new,
				Details: fmt.Sprintf("Added attribute %s=%q", key, newVal),
			})
		} else if !newExists {
			*diffs = append(*diffs, Diff{
				Type:    Removed,
				Path:    fmt.Sprintf("%s/@%s", path, key),
				Old:     old,
				Details: fmt.Sprintf("Removed attribute %s", key),
			})
		} else if oldVal != newVal {
			*diffs = append(*diffs, Diff{
				Type:    Modified,
				Path:    fmt.Sprintf("%s/@%s", path, key),
				Old:     old,
				New:     new,
				Details: fmt.Sprintf("Attribute %s changed: %q → %q", key, oldVal, newVal),
			})
		}
	}
}

func compareChildren(old, new *parser.Node, path string, diffs *[]Diff) {
	oldChildren := filterSignificant(old.Children)
	newChildren := filterSignificant(new.Children)

	// Simple comparison - compare by position
	maxLen := len(oldChildren)
	if len(newChildren) > maxLen {
		maxLen = len(newChildren)
	}

	for i := 0; i < maxLen; i++ {
		childPath := fmt.Sprintf("%s/%s[%d]", path, old.Tag, i)

		if i >= len(oldChildren) {
			*diffs = append(*diffs, Diff{
				Type:    Added,
				Path:    childPath,
				New:     newChildren[i],
				Details: fmt.Sprintf("Added child at position %d", i),
			})
		} else if i >= len(newChildren) {
			*diffs = append(*diffs, Diff{
				Type:    Removed,
				Path:    childPath,
				Old:     oldChildren[i],
				Details: fmt.Sprintf("Removed child at position %d", i),
			})
		} else {
			compareNodes(oldChildren[i], newChildren[i], childPath, diffs)
		}
	}
}

func filterSignificant(nodes []*parser.Node) []*parser.Node {
	var result []*parser.Node
	for _, n := range nodes {
		// Skip empty text nodes
		if n.Type == parser.NodeText && strings.TrimSpace(n.Text) == "" {
			continue
		}
		// Skip comments
		if n.Type == parser.NodeComment {
			continue
		}
		result = append(result, n)
	}
	return result
}

// Summary provides a summary of differences.
type Summary struct {
	Total    int
	Added    int
	Removed  int
	Modified int
}

// GetSummary returns a summary of the diffs.
func GetSummary(diffs []Diff) Summary {
	s := Summary{Total: len(diffs)}
	for _, d := range diffs {
		switch d.Type {
		case Added:
			s.Added++
		case Removed:
			s.Removed++
		case Modified:
			s.Modified++
		}
	}
	return s
}

// FormatText returns a human-readable text representation of diffs.
func FormatText(diffs []Diff) string {
	if len(diffs) == 0 {
		return "No differences found.\n"
	}

	var sb strings.Builder
	for _, d := range diffs {
		var prefix string
		switch d.Type {
		case Added:
			prefix = "+"
		case Removed:
			prefix = "-"
		case Modified:
			prefix = "~"
		default:
			prefix = " "
		}
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", prefix, d.Path, d.Details))
	}

	summary := GetSummary(diffs)
	sb.WriteString(fmt.Sprintf("\nSummary: %d changes (%d added, %d removed, %d modified)\n",
		summary.Total, summary.Added, summary.Removed, summary.Modified))

	return sb.String()
}

// FormatJSON returns a JSON representation of diffs.
func FormatJSON(diffs []Diff) string {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, d := range diffs {
		sb.WriteString(fmt.Sprintf("  {\n"))
		sb.WriteString(fmt.Sprintf("    \"type\": \"%s\",\n", d.Type))
		sb.WriteString(fmt.Sprintf("    \"path\": %q,\n", d.Path))
		sb.WriteString(fmt.Sprintf("    \"details\": %q\n", d.Details))
		if i < len(diffs)-1 {
			sb.WriteString("  },\n")
		} else {
			sb.WriteString("  }\n")
		}
	}
	sb.WriteString("]\n")
	return sb.String()
}
