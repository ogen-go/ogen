package ir

import (
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/ogen-go/ogen/internal/naming"
)

// markdownParser parses CommonMark with GitHub-flavored extensions (tables,
// strikethrough, autolinks, task lists).
var markdownParser = goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()

// renderMarkdown renders a CommonMark description into a slice of godoc comment
// lines (without the leading "// "), wrapping paragraphs at limit.
//
// The output uses Go doc comment conventions (see go/doc/comment): headings are
// emitted as "# Heading", links become "[text]" references with definitions
// collected at the end, lists and code blocks are indented as preformatted text.
func renderMarkdown(s string, limit int) []string {
	src := []byte(s)
	r := &mdRenderer{
		source:    src,
		limit:     limit,
		defs:      map[string]string{},
		firstPara: -1,
		lastPara:  -1,
	}
	r.renderChildren(markdownParser.Parse(text.NewReader(src)), "")

	// Trim trailing blank lines.
	for len(r.out) > 0 && r.out[len(r.out)-1] == "" {
		r.out = r.out[:len(r.out)-1]
	}

	// Capitalize the first paragraph line, matching plain-text doc conventions.
	if r.firstPara >= 0 {
		r.out[r.firstPara] = naming.Capitalize(r.out[r.firstPara])
	}
	// Ensure the description ends with a period, but only when the last
	// paragraph is the final block (not followed by a list, code block, etc.).
	if r.lastPara >= 0 && r.lastPara == len(r.out)-1 {
		if last := r.out[r.lastPara]; len(last) > 0 && last[len(last)-1] != '.' {
			r.out[r.lastPara] = last + "."
		}
	}

	// Append collected link definitions, e.g. "[text]: https://example.com".
	if len(r.links) > 0 {
		if len(r.out) > 0 {
			r.out = append(r.out, "")
		}
		for _, l := range r.links {
			r.out = append(r.out, "["+l.label+"]: "+l.url)
		}
	}

	return r.out
}

type mdLinkDef struct {
	label string
	url   string
}

type mdRenderer struct {
	source []byte
	limit  int

	out   []string
	links []mdLinkDef
	defs  map[string]string // label -> url, to dedupe definitions

	// firstPara and lastPara hold the indices in out of the first and last
	// paragraph lines, or -1 if there are none.
	firstPara int
	lastPara  int
}

func (r *mdRenderer) renderChildren(n ast.Node, indent string) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		r.renderBlock(c, indent)
	}
}

func (r *mdRenderer) renderBlock(n ast.Node, indent string) {
	switch node := n.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		r.ensureBlank()
		r.emitWrapped(r.inlineText(n), indent, true)
	case *ast.Heading:
		r.ensureBlank()
		// Go doc comments only recognize a single heading level ("# Heading").
		r.appendLine(indent + "# " + restoreSpaces(collapseSpaces(r.inlineText(n))))
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		r.ensureBlank()
		r.renderCode(n, indent)
	case *ast.List:
		r.renderList(node, indent)
	case *ast.Blockquote:
		// Go doc has no blockquote; render the contents as preformatted text.
		r.ensureBlank()
		r.renderChildren(n, indent+"\t")
	case *extast.Table:
		r.ensureBlank()
		r.renderTable(node, indent)
	case *ast.ThematicBreak:
		// Nothing sensible to render in godoc.
	default:
		r.renderChildren(n, indent)
	}
}

func (r *mdRenderer) renderList(n *ast.List, indent string) {
	r.ensureBlank()
	num := n.Start
	if num == 0 {
		num = 1
	}
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		var marker string
		if n.IsOrdered() {
			marker = strconv.Itoa(num) + ". "
			num++
		} else {
			marker = " - "
		}
		r.renderListItem(item, indent, marker)
	}
}

func (r *mdRenderer) renderListItem(item ast.Node, indent, marker string) {
	contIndent := indent + strings.Repeat(" ", len(marker))
	start := len(r.out)
	for c := item.FirstChild(); c != nil; c = c.NextSibling() {
		switch node := c.(type) {
		case *ast.Paragraph, *ast.TextBlock:
			r.emitWrapped(r.inlineText(c), contIndent, false)
		case *ast.List:
			r.renderList(node, contIndent)
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			r.ensureBlank()
			r.renderCode(c, contIndent)
		default:
			r.renderBlock(c, contIndent)
		}
	}
	// Replace the continuation indent of the item's first line with the marker.
	if len(r.out) > start {
		r.out[start] = indent + marker + strings.TrimPrefix(r.out[start], contIndent)
	}
}

func (r *mdRenderer) renderCode(n ast.Node, indent string) {
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		line := strings.TrimRight(string(seg.Value(r.source)), "\n")
		// Tab indentation marks the block as preformatted in godoc.
		r.out = append(r.out, indent+"\t"+line)
	}
}

func (r *mdRenderer) renderTable(n *extast.Table, indent string) {
	var rows [][]string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *extast.TableHeader, *extast.TableRow:
			rows = append(rows, r.tableRow(c))
		}
	}
	if len(rows) == 0 {
		return
	}

	cols := 0
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	widths := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	format := func(row []string) string {
		parts := make([]string, cols)
		for i := 0; i < cols; i++ {
			var cell string
			if i < len(row) {
				cell = row[i]
			}
			parts[i] = cell + strings.Repeat(" ", widths[i]-len(cell))
		}
		return strings.TrimRight(strings.Join(parts, " | "), " ")
	}

	// Tables are not supported by godoc, so render them as preformatted text.
	r.out = append(r.out, indent+"\t"+format(rows[0]))
	seps := make([]string, cols)
	for i := range seps {
		seps[i] = strings.Repeat("-", widths[i])
	}
	r.out = append(r.out, indent+"\t"+strings.Join(seps, "-+-"))
	for _, row := range rows[1:] {
		r.out = append(r.out, indent+"\t"+format(row))
	}
}

func (r *mdRenderer) tableRow(n ast.Node) (cells []string) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if cell, ok := c.(*extast.TableCell); ok {
			cells = append(cells, restoreSpaces(collapseSpaces(r.inlineText(cell))))
		}
	}
	return cells
}

// nbsp is a placeholder used to mark spaces inside atomic tokens (link labels,
// code spans) so they survive word wrapping. It is restored to a regular space
// once wrapping is done.
const nbsp = "\x00"

// atomicToken protects the spaces inside s from word wrapping.
func atomicToken(s string) string {
	return strings.ReplaceAll(s, " ", nbsp)
}

// restoreSpaces undoes atomicToken.
func restoreSpaces(s string) string {
	return strings.ReplaceAll(s, nbsp, " ")
}

// collapseSpaces collapses runs of whitespace into single spaces.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// inlineText renders the inline content of n into a single string, preserving
// the original spacing. Link labels and code spans have their internal spaces
// replaced with nbsp so word wrapping never splits them; the markers are
// restored by emitWrapped.
func (r *mdRenderer) inlineText(n ast.Node) string {
	var b strings.Builder
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch node := c.(type) {
			case *ast.Text:
				b.Write(node.Value(r.source))
				// Treat soft and hard line breaks as spaces; the paragraph is
				// re-wrapped anyway.
				if node.SoftLineBreak() || node.HardLineBreak() {
					b.WriteByte(' ')
				}
			case *ast.String:
				b.Write(node.Value)
			case *ast.CodeSpan:
				b.WriteString(atomicToken("`" + r.collectText(node) + "`"))
			case *ast.AutoLink:
				b.WriteString(atomicToken(string(node.URL(r.source))))
			case *ast.Link:
				b.WriteString(atomicToken(r.linkToken(r.collectText(node), string(node.Destination))))
			case *ast.Image:
				b.WriteString(atomicToken(r.linkToken(r.collectText(node), string(node.Destination))))
			default:
				// Emphasis, strikethrough, raw HTML, task list checkboxes, etc.:
				// drop the markup and keep the inner text.
				walk(c)
			}
		}
	}
	walk(n)
	return b.String()
}

// linkToken returns the inline token for a link with the given label and URL,
// recording a link definition to be emitted at the end of the comment.
func (r *mdRenderer) linkToken(label, url string) string {
	label = strings.Join(strings.Fields(label), " ")
	url = strings.TrimSpace(url)
	switch {
	case url == "":
		return label
	case label == "":
		label = url
	}
	// Brackets in the label would break the [label] reference syntax.
	label = strings.NewReplacer("[", "", "]", "").Replace(label)

	if existing, ok := r.defs[label]; ok {
		if existing == url {
			return "[" + label + "]"
		}
		// Conflicting definition for the same label: inline the URL instead.
		return label + " (" + url + ")"
	}
	r.defs[label] = url
	r.links = append(r.links, mdLinkDef{label: label, url: url})
	return "[" + label + "]"
}

// collectText returns the concatenated text content of n.
func (r *mdRenderer) collectText(n ast.Node) string {
	var b strings.Builder
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch node := c.(type) {
			case *ast.Text:
				b.Write(node.Value(r.source))
			case *ast.String:
				b.Write(node.Value)
			default:
				walk(c)
			}
		}
	}
	walk(n)
	return b.String()
}

// emitWrapped wraps s into lines prefixed with indent and appends them. Tokens
// are split on whitespace; nbsp markers inside atomic tokens are restored to
// spaces afterwards. When isPara is true, the produced lines are tracked for
// capitalization and the trailing period.
func (r *mdRenderer) emitWrapped(s, indent string, isPara bool) {
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return
	}
	limit := max(r.limit-len(indent), 1)

	var line strings.Builder
	flush := func() {
		if line.Len() == 0 {
			return
		}
		r.appendLine(indent + restoreSpaces(line.String()))
		if isPara {
			if r.firstPara < 0 {
				r.firstPara = len(r.out) - 1
			}
			r.lastPara = len(r.out) - 1
		}
		line.Reset()
	}
	for _, tok := range tokens {
		switch {
		case line.Len() == 0:
			line.WriteString(tok)
		case line.Len()+1+len(tok) > limit:
			flush()
			line.WriteString(tok)
		default:
			line.WriteByte(' ')
			line.WriteString(tok)
		}
	}
	flush()
}

func (r *mdRenderer) ensureBlank() {
	if len(r.out) > 0 && r.out[len(r.out)-1] != "" {
		r.out = append(r.out, "")
	}
}

func (r *mdRenderer) appendLine(s string) {
	r.out = append(r.out, strings.TrimRight(s, " "))
}
