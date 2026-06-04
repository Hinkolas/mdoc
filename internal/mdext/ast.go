package mdext

import (
	"github.com/hinkolas/mdoc/internal/document"
	gast "github.com/yuin/goldmark/ast"
)

// Directive is a `:::name [arg] … :::` block. The block parser fills
// Name/Arg/Options; the AST transformer fills Headings (for name=="toc") or Bib
// (for name=="bibliography") so the renderer can emit them without a
// parser.Context.
type Directive struct {
	gast.BaseBlock
	Name     string
	Arg      string // trailing token on the open line, e.g. `:::page cover`
	Options  map[string]string
	Headings []HeadingEntry
	Bib      []BibEntry
}

// KindDirective is the NodeKind of a Directive node.
var KindDirective = gast.NewNodeKind("Directive")

// Kind implements ast.Node.Kind.
func (n *Directive) Kind() gast.NodeKind { return KindDirective }

// Dump implements ast.Node.Dump.
func (n *Directive) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"Name": n.Name}, nil)
}

// NewDirective returns a Directive with the given name.
func NewDirective(name string) *Directive {
	return &Directive{Name: name, Options: map[string]string{}}
}

// HeadingEntry is one collected heading, used to build a table of contents.
type HeadingEntry struct {
	Level  int
	Number string // "2.1" / "A.1"; empty when the heading is {.unnumbered}
	Title  string // plain text, without the number
	ID     string
}

// BibEntry is one numbered, cited reference, used to build a bibliography.
type BibEntry struct {
	Number int
	Key    string
	Ref    document.Reference
}

// Citation is an inline `[@key]` (optionally `[@key, locator]`). The transformer
// fills Number/RefID/Resolved.
type Citation struct {
	gast.BaseInline
	Key      string
	Locator  string
	Number   int
	RefID    string
	Resolved bool
}

// KindCitation is the NodeKind of a Citation node.
var KindCitation = gast.NewNodeKind("Citation")

// Kind implements ast.Node.Kind.
func (n *Citation) Kind() gast.NodeKind { return KindCitation }

// Dump implements ast.Node.Dump.
func (n *Citation) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"Key": n.Key}, nil)
}

// NewCitation returns a Citation for the given key and optional locator.
func NewCitation(key, locator string) *Citation {
	return &Citation{Key: key, Locator: locator}
}

// SecNum is the section number injected as a numbered heading's first inline
// child, e.g. the "2.1" in "<h2>2.1 Title</h2>".
type SecNum struct {
	gast.BaseInline
	Num string
}

// KindSecNum is the NodeKind of a SecNum node.
var KindSecNum = gast.NewNodeKind("SecNum")

// Kind implements ast.Node.Kind.
func (n *SecNum) Kind() gast.NodeKind { return KindSecNum }

// Dump implements ast.Node.Dump.
func (n *SecNum) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"Num": n.Num}, nil)
}

// NewSecNum returns a SecNum carrying the given number text.
func NewSecNum(num string) *SecNum { return &SecNum{Num: num} }

// Matter is a document region (front matter / main matter / appendix). The
// transformer creates it by wrapping the nodes between two matter markers
// (`:::frontmatter` / `:::mainmatter` / `:::appendix`); it renders as a
// `<div class="mdoc-matter-<kind>">` the theme can style and break on.
type Matter struct {
	gast.BaseBlock
	Region string // "front" | "main" | "appendix"
}

// KindMatter is the NodeKind of a Matter node.
var KindMatter = gast.NewNodeKind("Matter")

// Kind implements ast.Node.Kind.
func (n *Matter) Kind() gast.NodeKind { return KindMatter }

// Dump implements ast.Node.Dump.
func (n *Matter) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"Region": n.Region}, nil)
}

// NewMatter returns a Matter region of the given kind.
func NewMatter(region string) *Matter { return &Matter{Region: region} }

// matterMarkers maps the directive name of a matter marker to its region kind.
var matterMarkers = map[string]string{
	"frontmatter": "front",
	"mainmatter":  "main",
	"appendix":    "appendix",
}
