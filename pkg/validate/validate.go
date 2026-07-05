// Package validate provides HTML validation and accessibility analysis.
package validate

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
)

// Severity indicates the severity of a validation issue.
type Severity int

const (
	Info Severity = iota
	Warning
	Error
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// Issue represents a validation issue found in HTML.
type Issue struct {
	Severity Severity
	Category string
	Message  string
	Line     int
	Node     *parser.Node
}

// Validator validates HTML documents.
type Validator struct {
	doc        *parser.Document
	issues     []Issue
	checkRules bool
}

// New creates a new Validator for the given document.
func New(doc *parser.Document) *Validator {
	return &Validator{doc: doc}
}

// Validate performs all validation checks and returns issues.
func (v *Validator) Validate() []Issue {
	v.issues = nil

	v.checkDocumentStructure()
	v.checkHeadingHierarchy()
	v.checkImages()
	v.checkLinks()
	v.checkForms()
	v.checkAccessibility()
	v.checkSecurity()
	v.checkSEO()

	return v.issues
}

func (v *Validator) addIssue(severity Severity, category, message string, node *parser.Node) {
	v.issues = append(v.issues, Issue{
		Severity: severity,
		Category: category,
		Message:  message,
		Node:     node,
	})
}

// checkDocumentStructure validates basic document structure.
func (v *Validator) checkDocumentStructure() {
	if v.doc.Root == nil {
		v.addIssue(Error, "structure", "Document has no root element", nil)
		return
	}

	// Check for doctype
	hasDoctype := false
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		if n.Type == parser.NodeDoctype {
			hasDoctype = true
		}
	})
	if !hasDoctype {
		v.addIssue(Warning, "structure", "Missing <!DOCTYPE html> declaration", nil)
	}

	// Check for html element
	hasHTML := false
	for _, child := range v.doc.Root.Children {
		if child.Tag == "html" {
			hasHTML = true
			break
		}
	}
	if !hasHTML {
		v.addIssue(Warning, "structure", "Missing <html> element", nil)
	}

	// Check for head and body
	hasHead := false
	hasBody := false
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		if n.Tag == "head" {
			hasHead = true
		}
		if n.Tag == "body" {
			hasBody = true
		}
	})
	if !hasHead {
		v.addIssue(Info, "structure", "Missing <head> element", nil)
	}
	if !hasBody {
		v.addIssue(Info, "structure", "Missing <body> element", nil)
	}

	// Check for title
	if v.doc.Title == "" {
		v.addIssue(Warning, "structure", "Missing <title> element", nil)
	}

	// Check for charset
	if _, ok := v.doc.Meta["charset"]; !ok {
		if _, ok := v.doc.Meta["Content-Type"]; !ok {
			v.addIssue(Warning, "structure", "Missing charset declaration", nil)
		}
	}

	// Check for viewport
	if _, ok := v.doc.Meta["viewport"]; !ok {
		v.addIssue(Info, "structure", "Missing viewport meta tag (recommended for responsive design)", nil)
	}
}

// checkHeadingHierarchy validates heading order.
func (v *Validator) checkHeadingHierarchy() {
	headings := v.doc.Headings
	if len(headings) == 0 {
		v.addIssue(Warning, "accessibility", "No headings found in document", nil)
		return
	}

	// Check for h1
	hasH1 := false
	for _, h := range headings {
		if h.Level == 1 {
			if hasH1 {
				v.addIssue(Warning, "structure", "Multiple <h1> elements found", h.Node)
			}
			hasH1 = true
			break
		}
	}
	if !hasH1 {
		v.addIssue(Warning, "structure", "No <h1> element found", nil)
	}

	// Check for skipped levels
	for i := 1; i < len(headings); i++ {
		if headings[i].Level > headings[i-1].Level+1 {
			v.addIssue(Warning, "structure",
				fmt.Sprintf("Heading level skipped: h%d after h%d", headings[i].Level, headings[i-1].Level),
				headings[i].Node)
		}
	}
}

// checkImages validates image accessibility.
func (v *Validator) checkImages() {
	images := v.doc.Images
	for _, img := range images {
		if img.Alt == "" {
			v.addIssue(Warning, "accessibility", fmt.Sprintf("Image missing alt text: %s", img.SRC), nil)
		}
		if img.SRC == "" {
			v.addIssue(Error, "structure", "Image with empty src attribute", nil)
		}
	}
}

// checkLinks validates link references.
func (v *Validator) checkLinks() {
	for _, link := range v.doc.Links {
		if link.Type == "a" {
			if link.HREF == "" {
				v.addIssue(Warning, "structure", fmt.Sprintf("Link with empty href: %s", link.Text), nil)
			} else if link.HREF == "#" {
				v.addIssue(Info, "structure", fmt.Sprintf("Placeholder link (#): %s", link.Text), nil)
			}
		}
	}
}

// checkForms validates form structure.
func (v *Validator) checkForms() {
	for _, form := range v.doc.Forms {
		if form.Action == "" {
			v.addIssue(Info, "structure", "Form with empty action attribute", nil)
		}
		if form.Method == "" {
			v.addIssue(Info, "structure", "Form with empty method attribute (defaults to GET)", nil)
		}
	}
}

// checkAccessibility checks for common accessibility issues.
func (v *Validator) checkAccessibility() {
	// Check for lang attribute on html
	langFound := false
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		if n.Tag == "html" {
			if _, ok := n.Attrs["lang"]; ok {
				langFound = true
			}
		}
	})
	if !langFound {
		v.addIssue(Warning, "accessibility", "Missing lang attribute on <html> element", nil)
	}

	// Check for images without alt
	for _, img := range v.doc.Images {
		if img.Alt == "" {
			v.addIssue(Warning, "accessibility", "Image without alt attribute", nil)
		}
	}

	// Check for links without accessible text
	for _, link := range v.doc.Links {
		if link.Type == "a" && link.Text == "" && link.HREF != "" {
			// Check if there's an image with alt text
			hasImageAlt := false
			parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
				if n.Tag == "a" && n.Attrs["href"] == link.HREF {
					for _, child := range n.Children {
						if child.Tag == "img" && child.Attrs["alt"] != "" {
							hasImageAlt = true
						}
					}
				}
			})
			if !hasImageAlt {
				v.addIssue(Warning, "accessibility",
					fmt.Sprintf("Link without accessible text: %s", link.HREF), nil)
			}
		}
	}

	// Check for color contrast (basic check - would need actual CSS for full check)
	// This is a placeholder for future enhancement

	// Check for tab order
	// Check for ARIA attributes
}

// checkSecurity checks for common security issues.
func (v *Validator) checkSecurity() {
	// Check for inline scripts
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		if n.Tag == "script" && n.Attrs["src"] == "" {
			v.addIssue(Info, "security", "Inline script detected (consider CSP)", n)
		}
	})

	// Check for inline event handlers
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		for attr := range n.Attrs {
			if strings.HasPrefix(attr, "on") {
				v.addIssue(Warning, "security",
					fmt.Sprintf("Inline event handler detected: %s", attr), n)
			}
		}
	})

	// Check for javascript: URLs
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		if n.Tag == "a" {
			href := n.Attrs["href"]
			if strings.HasPrefix(href, "javascript:") {
				v.addIssue(Warning, "security", "javascript: URL in link", n)
			}
		}
	})

	// Check for forms without CSRF protection
	for _, form := range v.doc.Forms {
		if form.Method == "post" {
			hasCSRF := false
			for _, field := range form.Fields {
				if strings.Contains(strings.ToLower(field.Name), "csrf") ||
					strings.Contains(strings.ToLower(field.Name), "token") {
					hasCSRF = true
					break
				}
			}
			if !hasCSRF {
				v.addIssue(Info, "security", "POST form without obvious CSRF token", nil)
			}
		}
	}
}

// checkSEO checks for basic SEO best practices.
func (v *Validator) checkSEO() {
	// Check for meta description
	if _, ok := v.doc.Meta["description"]; !ok {
		v.addIssue(Info, "seo", "Missing meta description", nil)
	}

	// Check for Open Graph tags
	hasOG := false
	for key := range v.doc.Meta {
		if strings.HasPrefix(key, "og:") {
			hasOG = true
			break
		}
	}
	if !hasOG {
		v.addIssue(Info, "seo", "Missing Open Graph meta tags", nil)
	}

	// Check for canonical URL
	hasCanonical := false
	parser.WalkNodes(v.doc.Root, func(n *parser.Node) {
		if n.Tag == "link" && n.Attrs["rel"] == "canonical" {
			hasCanonical = true
		}
	})
	if !hasCanonical {
		v.addIssue(Info, "seo", "Missing canonical link tag", nil)
	}

	// Check title length
	if v.doc.Title != "" {
		if len(v.doc.Title) < 30 {
			v.addIssue(Info, "seo", "Title may be too short (< 30 characters)", nil)
		}
		if len(v.doc.Title) > 60 {
			v.addIssue(Info, "seo", "Title may be too long (> 60 characters)", nil)
		}
	}
}

// Summary returns a summary of validation results.
type Summary struct {
	TotalIssues  int
	Errors       int
	Warnings     int
	InfoMessages int
	ByCategory   map[string]int
}

// GetSummary returns a summary of all validation issues.
func (v *Validator) GetSummary() Summary {
	s := Summary{
		ByCategory: make(map[string]int),
	}

	for _, issue := range v.issues {
		s.TotalIssues++
		switch issue.Severity {
		case Error:
			s.Errors++
		case Warning:
			s.Warnings++
		case Info:
			s.InfoMessages++
		}
		s.ByCategory[issue.Category]++
	}

	return s
}
