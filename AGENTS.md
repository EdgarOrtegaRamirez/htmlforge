# AGENTS.md

## Project Overview

HtmlForge is a comprehensive HTML processing toolkit for Go. It provides a CLI tool and library API for parsing, converting, minifying, validating, and comparing HTML documents.

## Architecture

```
pkg/
├── parser/     # HTML parser with DOM tree (Node, Document, WalkNodes)
├── extract/    # Content extraction (links, images, headings, metadata)
├── convert/    # HTML→Markdown and HTML→Text conversion
├── minify/     # HTML minification (remove comments, collapse whitespace)
├── format/     # HTML pretty-printing (configurable indentation)
├── validate/   # HTML validation (accessibility, security, SEO)
└── compare/    # Semantic HTML diffing (element-level changes)

cmd/htmlforge/  # CLI entry point using cobra
tests/          # Integration tests
```

## Development

```bash
# Build
go build ./...

# Run tests
go test ./...

# Build CLI
go build -o htmlforge ./cmd/htmlforge

# Run CLI
./htmlforge --help
```

## Key Concepts

1. **DOM Tree**: HTML is parsed into a tree of `Node` objects with type, tag, attributes, and children
2. **Document**: Contains parsed metadata (title, meta, links, images, headings, tables, forms)
3. **WalkNodes**: Depth-first traversal function for visiting all nodes
4. **TextContent**: Concatenated text from all descendant text nodes
5. **FindBySelector**: Simple CSS selector matching (#id, .class, tag)

## Adding New Features

1. **New extractor**: Add method to `extract/extract.go`
2. **New converter**: Add to `convert/convert.go` (handle new tags in `convertElementToMarkdown`)
3. **New validator**: Add check method to `validate/validate.go`
4. **New CLI command**: Add to `cmd/htmlforge/main.go` using cobra

## Testing

- Tests are in `tests/htmlforge_test.go`
- Run `go test ./... -v` for verbose output
- Each package has its own test file for unit tests

## Common Pitfalls

- The parser preserves whitespace in text nodes — trim when needed
- Document nodes (type=2) are the root — handle them in all processors
- Self-closing tags (br, img, etc.) have no closing tag
- Comment text is preserved in `node.Text`
