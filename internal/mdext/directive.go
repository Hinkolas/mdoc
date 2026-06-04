package mdext

import (
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// directiveParser parses single-line `:::name [arg] [key=value …]` directives —
// toc, bibliography, page, and the matter markers (frontmatter / mainmatter /
// appendix). They are leaf blocks (no closing fence), so a marker never
// swallows the content that follows it. Options are inline:
//
//	:::toc depth=3
//	:::page cover
//	:::frontmatter
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
	node := NewDirective(fields[0])
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
	return parser.Close // single line
}

func (b *directiveParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {}

func (b *directiveParser) CanInterruptParagraph() bool { return true }

func (b *directiveParser) CanAcceptIndentedLine() bool { return false }
