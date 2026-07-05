package htmlforge_test

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/compare"
	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/convert"
	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/extract"
	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/format"
	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/minify"
	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
	"github.com/EdgarOrtegaRamirez/htmlforge/pkg/validate"
)

// ==================== Parser Tests ====================

func TestParseString(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr bool
	}{
		{"simple html", "<html><head><title>Test</title></head><body><p>Hello</p></body></html>", false},
		{"empty string", "", false},
		{"self-closing tags", "<html><body><br><img src=\"test.jpg\"><input type=\"text\"></body></html>", false},
		{"nested elements", "<html><body><div><span><a href=\"#\">Link</a></span></div></body></html>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parser.ParseString(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if doc == nil && !tt.wantErr {
				t.Error("ParseString() returned nil document")
			}
		})
	}
}

func TestDocumentTitle(t *testing.T) {
	doc, err := parser.ParseString("<html><head><title>My Page</title></head><body></body></html>")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "My Page" {
		t.Errorf("Title = %q, want %q", doc.Title, "My Page")
	}
}

func TestDocumentMeta(t *testing.T) {
	doc, err := parser.ParseString(`<html><head>
		<meta name="description" content="Test page">
		<meta property="og:title" content="OG Title">
	</head><body></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Meta["description"] != "Test page" {
		t.Errorf("Meta description = %q, want %q", doc.Meta["description"], "Test page")
	}
	if doc.Meta["og:title"] != "OG Title" {
		t.Errorf("Meta og:title = %q, want %q", doc.Meta["og:title"], "OG Title")
	}
}

func TestDocumentLinks(t *testing.T) {
	doc, err := parser.ParseString(`<html><body>
		<a href="https://example.com">Example</a>
		<a href="/about">About</a>
	</body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Links) != 2 {
		t.Errorf("Links count = %d, want 2", len(doc.Links))
	}
	if doc.Links[0].HREF != "https://example.com" {
		t.Errorf("First link HREF = %q, want %q", doc.Links[0].HREF, "https://example.com")
	}
}

func TestDocumentImages(t *testing.T) {
	doc, err := parser.ParseString(`<html><body>
		<img src="photo.jpg" alt="A photo">
		<img src="icon.png" alt="">
	</body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Images) != 2 {
		t.Errorf("Images count = %d, want 2", len(doc.Images))
	}
	if doc.Images[0].Alt != "A photo" {
		t.Errorf("First image alt = %q, want %q", doc.Images[0].Alt, "A photo")
	}
}

func TestDocumentHeadings(t *testing.T) {
	doc, err := parser.ParseString(`<html><body>
		<h1>Main Title</h1>
		<h2>Section 1</h2>
		<h3>Subsection</h3>
	</body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Headings) != 3 {
		t.Errorf("Headings count = %d, want 3", len(doc.Headings))
	}
	if doc.Headings[0].Level != 1 || doc.Headings[0].Text != "Main Title" {
		t.Errorf("First heading wrong")
	}
}

func TestDocumentTables(t *testing.T) {
	doc, err := parser.ParseString(`<html><body>
		<table>
			<caption>Test</caption>
			<tr><th>Name</th><th>Age</th></tr>
			<tr><td>Alice</td><td>30</td></tr>
		</table>
	</body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Tables) != 1 {
		t.Errorf("Tables count = %d, want 1", len(doc.Tables))
	}
	if doc.Tables[0].Caption != "Test" {
		t.Errorf("Table caption = %q, want %q", doc.Tables[0].Caption, "Test")
	}
}

func TestFindByID(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><div id="content">Hello</div></body></html>`)
	node := doc.FindByID("content")
	if node == nil {
		t.Fatal("FindByID returned nil")
	}
	if node.TextContent() != "Hello" {
		t.Errorf("TextContent = %q, want %q", node.TextContent(), "Hello")
	}
}

func TestFindByTag(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><p>A</p><p>B</p><div>C</div></body></html>`)
	nodes := doc.FindByTag("p")
	if len(nodes) != 2 {
		t.Errorf("FindByTag count = %d, want 2", len(nodes))
	}
}

func TestFindByClass(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><div class="card active">A</div><div class="card">B</div></body></html>`)
	nodes := doc.FindByClass("card")
	if len(nodes) != 2 {
		t.Errorf("FindByClass count = %d, want 2", len(nodes))
	}
}

func TestFindBySelector(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><div id="main">M</div><p>P</p></body></html>`)
	nodes := doc.FindBySelector("#main")
	if len(nodes) != 1 || nodes[0].TextContent() != "M" {
		t.Errorf("FindBySelector #main failed")
	}
	nodes = doc.FindBySelector("p")
	if len(nodes) != 1 {
		t.Errorf("FindBySelector p failed")
	}
}

func TestTextContent(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><h1>Title</h1><p>Hello <strong>World</strong></p></body></html>`)
	text := doc.Text
	if !strings.Contains(text, "Title") || !strings.Contains(text, "Hello World") {
		t.Errorf("Text = %q, missing expected content", text)
	}
}

func TestOuterHTML(t *testing.T) {
	doc, _ := parser.ParseString(`<div id="test">Hello</div>`)
	nodes := doc.FindByTag("div")
	html := nodes[0].OuterHTML()
	if !strings.Contains(html, "id=\"test\"") || !strings.Contains(html, "Hello") {
		t.Errorf("OuterHTML = %q", html)
	}
}

// ==================== Extract Tests ====================

func TestExtractText(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><h1>Hello</h1><p>World</p></body></html>`)
	ext := extract.New(doc)
	text := ext.Text(extract.DefaultTextOptions())
	if !strings.Contains(text, "Hello") || !strings.Contains(text, "World") {
		t.Errorf("Text = %q", text)
	}
}

func TestExtractLinks(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><a href="a">A</a><a href="b">B</a></body></html>`)
	ext := extract.New(doc)
	links := ext.Links()
	if len(links) != 2 {
		t.Errorf("Links = %d, want 2", len(links))
	}
}

func TestExtractHeadings(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><h1>A</h1><h2>B</h2></body></html>`)
	ext := extract.New(doc)
	headings := ext.Headings()
	if len(headings) != 2 {
		t.Errorf("Headings = %d, want 2", len(headings))
	}
}

func TestExtractHeadingTree(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><h1>A</h1><h2>B</h2><h2>C</h2></body></html>`)
	ext := extract.New(doc)
	tree := ext.HeadingTree()
	if tree == nil || len(tree.Children) != 1 {
		t.Error("HeadingTree failed")
	}
}

func TestExtractTitle(t *testing.T) {
	doc, _ := parser.ParseString(`<html><head><title>My Title</title></head><body></body></html>`)
	ext := extract.New(doc)
	if ext.Title() != "My Title" {
		t.Errorf("Title = %q, want %q", ext.Title(), "My Title")
	}
}

func TestExtractOpenGraph(t *testing.T) {
	doc, _ := parser.ParseString(`<html><head>
		<meta property="og:title" content="OG Title">
		<meta name="description" content="Regular">
	</head><body></body></html>`)
	ext := extract.New(doc)
	og := ext.OpenGraph()
	if len(og) != 1 || og["og:title"] != "OG Title" {
		t.Errorf("OpenGraph = %v", og)
	}
}

func TestExtractUniqueLinks(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body>
		<a href="a">First</a><a href="a">Second</a><a href="b">Third</a>
	</body></html>`)
	ext := extract.New(doc)
	unique := ext.UniqueLinks()
	if len(unique) != 2 {
		t.Errorf("UniqueLinks = %d, want 2", len(unique))
	}
}

func TestExtractImagesWithoutAlt(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body>
		<img src="a.jpg" alt="OK"><img src="b.jpg" alt="">
	</body></html>`)
	ext := extract.New(doc)
	noAlt := ext.ImagesWithoutAlt()
	if len(noAlt) != 1 {
		t.Errorf("ImagesWithoutAlt = %d, want 1", len(noAlt))
	}
}

func TestExtractBrokenAnchors(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body>
		<a href="valid">V</a><a>No href</a>
	</body></html>`)
	ext := extract.New(doc)
	broken := ext.BrokenAnchors()
	if len(broken) != 1 {
		t.Errorf("BrokenAnchors = %d, want 1", len(broken))
	}
}

// ==================== Convert Tests ====================

func TestToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		contains []string
	}{
		{"headings", "<html><body><h1>Title</h1></body></html>", []string{"# Title"}},
		{"bold", "<html><body><p><strong>Bold</strong></p></body></html>", []string{"**Bold**"}},
		{"italic", "<html><body><p><em>Italic</em></p></body></html>", []string{"*Italic*"}},
		{"link", "<html><body><a href=\"url\">Text</a></body></html>", []string{"[Text](url)"}},
		{"image", "<html><body><img src=\"img.jpg\" alt=\"Alt\"></body></html>", []string{"![Alt](img.jpg)"}},
		{"code", "<html><body><p><code>x</code></p></body></html>", []string{"`x`"}},
		{"list", "<html><body><ul><li>A</li></ul></body></html>", []string{"- A"}},
		{"ordered", "<html><body><ol><li>A</li></ol></body></html>", []string{"1. A"}},
		{"blockquote", "<html><body><blockquote>Q</blockquote></body></html>", []string{"> Q"}},
		{"hr", "<html><body><hr></body></html>", []string{"---"}},
		{"strikethrough", "<html><body><del>X</del></body></html>", []string{"~~X~~"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := parser.ParseString(tt.html)
			md := convert.ToMarkdown(doc)
			for _, s := range tt.contains {
				if !strings.Contains(md, s) {
					t.Errorf("Markdown should contain %q, got %q", s, md)
				}
			}
		})
	}
}

func TestToText(t *testing.T) {
	doc, _ := parser.ParseString(`<html><body><p>Hello <strong>World</strong></p><script>x</script></body></html>`)
	text := convert.ToText(doc)
	if !strings.Contains(text, "Hello World") {
		t.Errorf("Text = %q", text)
	}
	if strings.Contains(text, "x") {
		t.Errorf("Text should not contain script content")
	}
}

func TestMarkdownTable(t *testing.T) {
	html := `<html><body><table><tr><th>A</th></tr><tr><td>1</td></tr></table></body></html>`
	doc, _ := parser.ParseString(html)
	md := convert.ToMarkdown(doc)
	if !strings.Contains(md, "| A |") || !strings.Contains(md, "| 1 |") {
		t.Errorf("Markdown table = %q", md)
	}
}

// ==================== Minify Tests ====================

func TestMinifyHTML(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		contains    []string
		notContains []string
	}{
		{"removes comments", "<html><!-- c --><body><p>H</p></body></html>", []string{"<p>H</p>"}, []string{"<!--"}},
		{"collapses whitespace", "<html><body>  <p>   H   W   </p>  </body></html>", []string{"H W"}, nil},
		{"preserves structure", "<html><head><title>T</title></head><body></body></html>", []string{"<title>T</title>"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minify.MinifyHTML(tt.html)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Should contain %q, got %q", s, result)
				}
			}
			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("Should not contain %q, got %q", s, result)
				}
			}
		})
	}
}

func TestMinifyWithCustomConfig(t *testing.T) {
	m := minify.Default()
	m.RemoveComments = false
	result := m.Minify("<!-- keep --><p>Hi</p>")
	if !strings.Contains(result, "<!-- keep -->") {
		t.Errorf("Should keep comments, got %q", result)
	}
}

// ==================== Format Tests ====================

func TestFormatHTML(t *testing.T) {
	result := format.FormatHTML("<html><body><div><p>Hello</p></div></body></html>")
	if !strings.Contains(result, "  <div>") {
		t.Errorf("Should indent, got %q", result)
	}
}

func TestFormatWithIndentAttrs(t *testing.T) {
	f := format.Default()
	f.IndentAttrs = true
	result := f.Format(`<div id="a" class="b">X</div>`)
	lines := strings.Split(result, "\n")
	attrCount := 0
	for _, line := range lines {
		if strings.Contains(line, "id=") || strings.Contains(line, "class=") {
			attrCount++
		}
	}
	if attrCount < 2 {
		t.Errorf("Should have separate attr lines, got %q", result)
	}
}

func TestMinifyFunction(t *testing.T) {
	result := format.Minify("  <p>  Hello  </p>  ")
	if strings.Contains(result, "  ") {
		t.Errorf("Should collapse spaces, got %q", result)
	}
}

// ==================== Validate Tests ====================

func TestValidateValidDocument(t *testing.T) {
	html := `<!DOCTYPE html><html lang="en"><head>
		<meta charset="UTF-8"><meta name="viewport" content="width=device-width">
		<meta name="description" content="Desc"><title>Title</title>
	</head><body><h1>Title</h1><p>Text</p><img src="img.jpg" alt="Alt"></body></html>`

	doc, _ := parser.ParseString(html)
	v := validate.New(doc)
	issues := v.Validate()

	errors := 0
	for _, issue := range issues {
		if issue.Severity == validate.Error {
			errors++
		}
	}
	if errors > 0 {
		t.Errorf("Valid doc should have 0 errors, got %d", errors)
	}
}

func TestValidateMissingDoctype(t *testing.T) {
	doc, _ := parser.ParseString(`<html><head><title>T</title></head><body></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "DOCTYPE") {
			found = true
		}
	}
	if !found {
		t.Error("Should warn about missing DOCTYPE")
	}
}

func TestValidateMissingTitle(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html><head></head><body></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "title") {
			found = true
		}
	}
	if !found {
		t.Error("Should warn about missing title")
	}
}

func TestValidateImagesWithoutAlt(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>T</title></head>
		<body><img src="a.jpg"><img src="b.jpg" alt="OK"></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	count := 0
	for _, issue := range issues {
		if strings.Contains(issue.Message, "alt") {
			count++
		}
	}
	if count == 0 {
		t.Error("Should warn about missing alt")
	}
}

func TestValidateHeadingHierarchy(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>T</title></head>
		<body><h1>A</h1><h3>C</h3></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "skipped") || strings.Contains(issue.Message, "Skipped") {
			found = true
		}
	}
	if !found {
		t.Error("Should warn about skipped heading levels")
	}
}

func TestValidateInlineEventHandlers(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>T</title></head>
		<body><div onclick="alert('x')">Click</div></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "event handler") {
			found = true
		}
	}
	if !found {
		t.Error("Should warn about inline event handlers")
	}
}

func TestValidateJavascriptURLs(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>T</title></head>
		<body><a href="javascript:void(0)">Bad</a></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "javascript:") {
			found = true
		}
	}
	if !found {
		t.Error("Should warn about javascript: URLs")
	}
}

func TestValidateMissingLang(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>T</title></head>
		<body><p>Hello</p></body></html>`)
	v := validate.New(doc)
	issues := v.Validate()

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "lang") {
			found = true
		}
	}
	if !found {
		t.Error("Should warn about missing lang attribute")
	}
}

func TestGetSummary(t *testing.T) {
	doc, _ := parser.ParseString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>T</title></head>
		<body><h1>T</h1><img src="a.jpg"><a href="#">L</a></body></html>`)
	v := validate.New(doc)
	v.Validate()
	summary := v.GetSummary()
	if summary.TotalIssues == 0 {
		t.Error("Summary should have issues")
	}
}

// ==================== Compare Tests ====================

func TestCompareIdentical(t *testing.T) {
	html := `<html><body><p>Hello</p></body></html>`
	doc1, _ := parser.ParseString(html)
	doc2, _ := parser.ParseString(html)
	diffs := compare.Compare(doc1, doc2)
	if len(diffs) != 0 {
		t.Errorf("Identical docs should have 0 diffs, got %d", len(diffs))
	}
}

func TestCompareTextChange(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><p>Old</p></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><p>New</p></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	if len(diffs) == 0 {
		t.Error("Changed text should produce diffs")
	}
}

func TestCompareAddedElement(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><p>New</p></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	found := false
	for _, d := range diffs {
		if d.Type == compare.Added {
			found = true
		}
	}
	if !found {
		t.Error("Should detect added element")
	}
}

func TestCompareRemovedElement(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><p>Old</p></body></html>`)
	doc2, _ := parser.ParseString(`<html><body></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	found := false
	for _, d := range diffs {
		if d.Type == compare.Removed {
			found = true
		}
	}
	if !found {
		t.Error("Should detect removed element")
	}
}

func TestCompareAttributeChange(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><a href="old.html">Link</a></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><a href="new.html">Link</a></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	found := false
	for _, d := range diffs {
		if strings.Contains(d.Details, "href") {
			found = true
		}
	}
	if !found {
		t.Error("Should detect attribute change")
	}
}

func TestCompareTagChange(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><p>T</p></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><div>T</div></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	found := false
	for _, d := range diffs {
		if strings.Contains(d.Details, "Tag changed") {
			found = true
		}
	}
	if !found {
		t.Error("Should detect tag change")
	}
}

func TestGetSummaryCompare(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><p>Old</p></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><p>New</p></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	summary := compare.GetSummary(diffs)
	if summary.Total == 0 {
		t.Error("Summary should have total > 0")
	}
}

func TestFormatText(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><p>A</p></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><p>B</p></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	text := compare.FormatText(diffs)
	if !strings.Contains(text, "Summary") {
		t.Error("FormatText should contain summary")
	}
}

func TestFormatJSON(t *testing.T) {
	doc1, _ := parser.ParseString(`<html><body><p>A</p></body></html>`)
	doc2, _ := parser.ParseString(`<html><body><p>B</p></body></html>`)
	diffs := compare.Compare(doc1, doc2)
	json := compare.FormatJSON(diffs)
	if !strings.HasPrefix(json, "[") || !strings.HasSuffix(strings.TrimSpace(json), "]") {
		t.Error("FormatJSON should be valid JSON array")
	}
}

func TestFormatTextNoDiffs(t *testing.T) {
	html := `<html><body><p>S</p></body></html>`
	doc1, _ := parser.ParseString(html)
	doc2, _ := parser.ParseString(html)
	diffs := compare.Compare(doc1, doc2)
	text := compare.FormatText(diffs)
	if !strings.Contains(text, "No differences") {
		t.Error("Should report no differences")
	}
}
