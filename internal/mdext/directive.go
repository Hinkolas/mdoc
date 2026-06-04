package mdext

import (
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// containerDirectives are the `:::name … :::` directives that carry a markdown
// body (closed by a bare `:::` fence). Everything else is a single-line leaf.
var containerDirectives = map[string]bool{
	"figure": true,
	"table":  true,
}

// directiveParser parses `:::…` directives. Most are single-line leaf blocks —
// toc, bibliography, lof, lot, page, and the matter markers (frontmatter /
// mainmatter / appendix) — so a marker never swallows the content that follows
// it. Options are inline:
//
//	:::toc depth=3
//	:::page cover
//	:::frontmatter
//
// figure and table are containers: `:::figure #id` opens a block whose markdown
// body (image/table media plus a rich caption) runs until a closing `:::`.
type directiveParser struct{}

// NewDirectiveParser returns the `:::…` directive block parser.
func NewDirectiveParser() parser.BlockParser { return &directiveParser{} }

func (b *directiveParser) Trigger() []byte { return []byte{':'} }

func (b *directiveParser) Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockIndent()
	if pos < 0 {
		return nil, parser.NoChildren
	}
	i := pos
	for ; i < len(line) && line[i] == ':'; i++ {
	}
	if i-pos < 3 {
		return nil, parser.NoChildren
	}
	fields := strings.Fields(string(line[i:]))
	if len(fields) == 0 {
		return nil, parser.NoChildren
	}
	name := fields[0]

	if containerDirectives[name] {
		node := NewCaptioned(name)
		for _, f := range fields[1:] {
			if id, ok := strings.CutPrefix(f, "#"); ok {
				node.ID = id
			} else if k, v, ok := strings.Cut(f, "="); ok {
				node.Options[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		// Stop on the opener's newline (not past it) so goldmark opens no child
		// from this line; the body is parsed line by line via Continue.
		reader.AdvanceToEOL()
		return node, parser.HasChildren
	}

	node := NewDirective(name)
	for _, f := range fields[1:] {
		if k, v, ok := strings.Cut(f, "="); ok {
			node.Options[strings.TrimSpace(k)] = strings.TrimSpace(v)
		} else if node.Arg == "" {
			node.Arg = f
		}
	}
	reader.AdvanceToEOL()
	return node, parser.NoChildren
}

func (b *directiveParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	if _, ok := node.(*Captioned); ok {
		// A container runs until a bare `:::` fence. The body has no per-line
		// marker, so Continue | HasChildren tells goldmark to parse each body line
		// as a child block (image paragraphs, a table, the caption).
		line, _ := reader.PeekLine()
		if isCloseFence(line) {
			reader.AdvanceToEOL() // consume the fence, leaving its newline
			return parser.Close
		}
		return parser.Continue | parser.HasChildren
	}
	return parser.Close // leaf directive: single line
}

func (b *directiveParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {}

func (b *directiveParser) CanInterruptParagraph() bool { return true }

func (b *directiveParser) CanAcceptIndentedLine() bool { return false }

// isCloseFence reports whether a line is a bare `:::` (three or more colons,
// nothing else), the closing fence of a container directive.
func isCloseFence(line []byte) bool {
	s := strings.TrimSpace(string(line))
	if len(s) < 3 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ':' {
			return false
		}
	}
	return true
}
