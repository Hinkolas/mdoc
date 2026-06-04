// Package mdext is mdoc's goldmark extension. It adds markdown-native document
// apparatus: auto section numbering, a `:::toc` table of contents, `[@key]`
// citations, and a `:::bibliography` reference list — so the body stays
// markdown instead of hand-written HTML.
//
// The extension is constructed per render with the document's frontmatter (see
// New); a single AST transformer builds the document model and the node
// renderers emit a stable `mdoc-*` CSS-class contract that themes style.
package mdext

import (
	"github.com/hinkolas/mdoc/internal/document"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Config parameterises the extension with the document's frontmatter.
type Config struct {
	References []document.Reference
	Numbering  document.Numbering
}

type extender struct {
	cfg Config
}

// New returns a goldmark extender for the given document config. Build it per
// Convert so it sees that document's references and numbering.
func New(cfg Config) goldmark.Extender { return &extender{cfg: cfg} }

// Extend registers the directive block parser, the citation inline parser, the
// numbering/collection transformer, and the node renderers.
//
// Priorities: the citation inline parser runs ahead of goldmark's link (200)
// and footnote (101) parsers so it can claim `[@…`, while returning nil for
// everything else so links and footnotes still work.
func (e *extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewDirectiveParser(), 100),
		),
		parser.WithInlineParsers(
			util.Prioritized(NewCitationParser(), 100),
		),
		parser.WithASTTransformers(
			util.Prioritized(newTransformer(e.cfg), 100),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewNodeRenderer(), 100),
	))
}
