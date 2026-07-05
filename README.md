# HtmlForge

A comprehensive HTML processing toolkit for Go with a CLI tool and library API for parsing, converting, minifying, validating, and comparing HTML documents.

## Features

- **HTML Parser** — DOM tree with node traversal, text extraction, element finding
- **Content Extraction** — Links, images, headings, metadata, forms, tables
- **HTML→Markdown** — Convert HTML to clean Markdown with proper formatting
- **HTML→Text** — Extract plain text with structure preservation
- **HTML Minification** — Remove comments, collapse whitespace, optimize output
- **HTML Formatting** — Pretty-print with configurable indentation
- **HTML Validation** — Structural, accessibility, security, and SEO checks
- **HTML Diffing** — Semantic comparison with element-level change tracking

## Install

```bash
go install github.com/EdgarOrtegaRamirez/htmlforge/cmd/htmlforge@latest
```

## Quick Start

```bash
# Extract all content from HTML
htmlforge extract --what all page.html

# Convert to Markdown
htmlforge convert --to markdown page.html

# Minify HTML
htmlforge minify page.html

# Pretty-print HTML
htmlforge format page.html

# Validate HTML
htmlforge validate page.html

# Compare two HTML files
htmlforge diff old.html new.html

# Show document info
htmlforge info page.html
```

## Library API

```go
import (
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/parser"
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/extract"
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/convert"
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/minify"
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/format"
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/validate"
    "github.com/EdgarOrtegaRamirez/htmlforge/pkg/compare"
)

// Parse HTML
doc, _ := parser.ParseString("<html><body><p>Hello</p></body></html>")

// Extract content
ext := extract.New(doc)
fmt.Println(ext.Title())
fmt.Println(ext.Links())
fmt.Println(ext.Images())

// Convert to Markdown
md := convert.ToMarkdown(doc)

// Minify
m := minify.Default()
minified := m.Minify(html)

// Format
f := format.Default()
formatted := f.Format(html)

// Validate
v := validate.New(doc)
issues := v.Validate()

// Compare
diffs := compare.Compare(oldDoc, newDoc)
```

## Commands

| Command | Description |
|---------|-------------|
| `extract` | Extract content (text, links, images, headings, meta, og) |
| `convert` | Convert HTML to Markdown or plain text |
| `minify` | Remove comments and collapse whitespace |
| `format` | Pretty-print with indentation |
| `validate` | Check for accessibility, security, SEO issues |
| `diff` | Compare two HTML files semantically |
| `info` | Show document structure information |

## Architecture

```
pkg/
├── parser/     # HTML parser with DOM tree
├── extract/    # Content extraction (links, images, headings, etc.)
├── convert/    # HTML→Markdown and HTML→Text conversion
├── minify/     # HTML minification
├── format/     # HTML pretty-printing
├── validate/   # HTML validation and accessibility analysis
└── compare/    # Semantic HTML diffing
```

## License

MIT
