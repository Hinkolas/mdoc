package mdext

import (
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// citationParser parses inline `[@key]` citations and `[#id]` cross-references.
// It triggers on '[' but returns nil for anything that isn't `[@…` or `[#…`, so
// ordinary links (`[text](url)`) and footnotes (`[^id]`) fall through to their
// own parsers. Register it at a higher priority (lower number) than those.
type citationParser struct{}

// NewCitationParser returns the `[@key]` / `[#id]` inline parser.
func NewCitationParser() parser.InlineParser { return &citationParser{} }

func (s *citationParser) Trigger() []byte { return []byte{'['} }

func (s *citationParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
	if len(line) < 4 || line[0] != '[' {
		return nil
	}
	switch line[1] {
	case '@':
		return parseCitation(line, block)
	case '#':
		return parseXref(line, block)
	default:
		return nil
	}
}

// closeBracket returns the index of the closing ']' on the same line, or -1.
// Citations and cross-references don't span lines or nest.
func closeBracket(line []byte) int {
	for i := 2; i < len(line); i++ {
		switch line[i] {
		case ']':
			return i
		case '\n':
			return -1
		}
	}
	return -1
}

// parseCitation parses `[@key]` / `[@key, locator]`.
func parseCitation(line []byte, block text.Reader) gast.Node {
	end := closeBracket(line)
	if end < 0 {
		return nil
	}
	inner := strings.TrimSpace(string(line[2:end]))
	block.Advance(end + 1) // consume "[@" + inner + "]"

	key, locator, _ := strings.Cut(inner, ",")
	key = strings.TrimSpace(key)
	locator = strings.TrimSpace(locator)
	if key == "" {
		return nil
	}
	return NewCitation(key, locator)
}

// parseXref parses `[#id]` (the target's number) and `[#id page]` (its page
// number). It declines `[#x](url)` / `[#x][ref]` so those stay markdown links.
func parseXref(line []byte, block text.Reader) gast.Node {
	end := closeBracket(line)
	if end < 0 {
		return nil
	}
	if end+1 < len(line) && (line[end+1] == '(' || line[end+1] == '[') {
		return nil
	}
	fields := strings.FieldsFunc(string(line[2:end]), func(r rune) bool {
		return r == ' ' || r == '\t' || r == ','
	})
	if len(fields) == 0 {
		return nil
	}
	mode := "num"
	if len(fields) > 1 && fields[1] == "page" {
		mode = "page"
	}
	block.Advance(end + 1) // consume "[#" + inner + "]"
	return NewXref(fields[0], mode)
}
