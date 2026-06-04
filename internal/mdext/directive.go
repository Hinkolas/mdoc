package mdext

import (
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// directiveParser parses `:::name … :::` fenced blocks. It mirrors goldmark's
// fenced-code-block parser (parser/fcode_block.go): the open line carries the
// directive name, inner lines are collected verbatim, and `:::` on its own line
// closes it. Inner lines are parsed into key:value options in Close.
type directiveParser struct{}

// NewDirectiveParser returns the `:::…:::` block parser.
func NewDirectiveParser() parser.BlockParser { return &directiveParser{} }

type directiveData struct {
	length int // number of colons in the opening fence
	node   *Directive
}

var directiveInfoKey = parser.NewContextKey()

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
	fenceLength := i - pos
	if fenceLength < 3 {
		return nil, parser.NoChildren
	}
	name := strings.TrimSpace(string(line[i:]))
	// Keep only the first token as the name; ignore any trailing text on the
	// open line (options live on their own lines inside the fence).
	if sp := strings.IndexAny(name, " \t"); sp >= 0 {
		name = name[:sp]
	}
	if name == "" {
		return nil, parser.NoChildren
	}
	node := NewDirective(name)
	pc.Set(directiveInfoKey, &directiveData{length: fenceLength, node: node})
	return node, parser.NoChildren
}

func (b *directiveParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	fdata := pc.Get(directiveInfoKey).(*directiveData)

	// A line of >= the opening number of colons, with nothing else, closes.
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w < 4 {
		i := pos
		for ; i < len(line) && line[i] == ':'; i++ {
		}
		if i-pos >= fdata.length && util.IsBlank(line[i:]) {
			reader.AdvanceToEOL()
			return parser.Close
		}
	}
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (b *directiveParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {
	d := node.(*Directive)
	src := reader.Source()
	lines := d.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		raw := strings.TrimSpace(string(seg.Value(src)))
		if raw == "" {
			continue
		}
		if k, v, ok := splitOption(raw); ok {
			d.Options[k] = v
		}
	}
	if fdata := pc.Get(directiveInfoKey); fdata != nil && fdata.(*directiveData).node == d {
		pc.Set(directiveInfoKey, nil)
	}
}

func (b *directiveParser) CanInterruptParagraph() bool { return true }

func (b *directiveParser) CanAcceptIndentedLine() bool { return false }

// splitOption parses a `key: value` (or `key = value`) option line.
func splitOption(s string) (key, value string, ok bool) {
	idx := strings.IndexAny(s, ":=")
	if idx <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(s[:idx])
	value = strings.TrimSpace(s[idx+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}
