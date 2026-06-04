package mdext

import (
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// citationParser parses inline `[@key]` (and `[@key, locator]`) citations. It
// triggers on '[' but returns nil for anything that isn't `[@…`, so ordinary
// links (`[text](url)`) and footnotes (`[^id]`) fall through to their own
// parsers. Register it at a higher priority (lower number) than those.
type citationParser struct{}

// NewCitationParser returns the `[@key]` inline parser.
func NewCitationParser() parser.InlineParser { return &citationParser{} }

func (s *citationParser) Trigger() []byte { return []byte{'['} }

func (s *citationParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
	if len(line) < 4 || line[0] != '[' || line[1] != '@' {
		return nil
	}
	// Find the closing ']' on the same line (citations don't span lines or nest).
	end := -1
	for i := 2; i < len(line); i++ {
		if line[i] == ']' {
			end = i
			break
		}
		if line[i] == '\n' {
			break
		}
	}
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
