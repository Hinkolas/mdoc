package mdext

import (
	"github.com/hinkolas/mdoc/internal/document"
	gast "github.com/yuin/goldmark/ast"
)

// Directive is a single-line `:::name [arg] [key=value …]` leaf block (toc,
// bibliography, lof, lot, page, and the matter markers). The block parser fills
// Name/Arg/Options; the AST transformer fills Headings (name=="toc"), Bib
// (name=="bibliography"), or Entries (name=="lof"/"lot") so the renderer can
// emit them without a parser.Context.
type Directive struct {
	gast.BaseBlock
	Name     string
	Arg      string // trailing token on the open line, e.g. `:::page cover`
	Options  map[string]string
	Headings []HeadingEntry
	Bib      []BibEntry
	Entries  []CaptionEntry
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

// CaptionEntry is one collected figure or table, used to build a list of figures
// (`:::lof`) or tables (`:::lot`).
type CaptionEntry struct {
	Number string // "2.1" / "A.1"
	Title  string // plain caption text (falls back to the image alt)
	ID     string
}

// Captioned is a `:::figure … :::` or `:::table … :::` block. Its body is normal
// markdown: image-bearing paragraphs (or a table) are the media, the remaining
// text paragraphs are the caption. The transformer numbers it, separates media
// from caption (a Caption child), and injects the "Abbildung 2.1" label.
type Captioned struct {
	gast.BaseBlock
	Variant string // "figure" | "table"
	ID      string
	Number  string
	Options map[string]string
}

// KindCaptioned is the NodeKind of a Captioned node.
var KindCaptioned = gast.NewNodeKind("Captioned")

// Kind implements ast.Node.Kind.
func (n *Captioned) Kind() gast.NodeKind { return KindCaptioned }

// Dump implements ast.Node.Dump.
func (n *Captioned) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"Variant": n.Variant, "ID": n.ID}, nil)
}

// NewCaptioned returns a Captioned of the given variant.
func NewCaptioned(variant string) *Captioned {
	return &Captioned{Variant: variant, Options: map[string]string{}}
}

// Caption holds the caption of a Captioned block. It carries inline children
// (the injected label followed by the author's rich caption text) and renders
// as a `<figcaption>`.
type Caption struct {
	gast.BaseBlock
}

// KindCaption is the NodeKind of a Caption node.
var KindCaption = gast.NewNodeKind("Caption")

// Kind implements ast.Node.Kind.
func (n *Caption) Kind() gast.NodeKind { return KindCaption }

// Dump implements ast.Node.Dump.
func (n *Caption) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// NewCaption returns an empty Caption.
func NewCaption() *Caption { return &Caption{} }

// CaptionLabel is the "Abbildung 2.1" / "Tabelle 2.1" lead injected as a
// caption's first inline child. Class selects the per-variant CSS class. The
// field is Label (not Text) to avoid shadowing ast.Node's Text method.
type CaptionLabel struct {
	gast.BaseInline
	Label string
	Class string
}

// KindCaptionLabel is the NodeKind of a CaptionLabel node.
var KindCaptionLabel = gast.NewNodeKind("CaptionLabel")

// Kind implements ast.Node.Kind.
func (n *CaptionLabel) Kind() gast.NodeKind { return KindCaptionLabel }

// Dump implements ast.Node.Dump.
func (n *CaptionLabel) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"Label": n.Label}, nil)
}

// NewCaptionLabel returns a CaptionLabel with the given text and CSS class.
func NewCaptionLabel(label, class string) *CaptionLabel {
	return &CaptionLabel{Label: label, Class: class}
}

// Xref is an inline cross-reference `[#id]` (the target element's number) or
// `[#id page]` (its page number, resolved by the theme via target-counter). The
// transformer fills Number/Resolved.
type Xref struct {
	gast.BaseInline
	ID       string
	Mode     string // "num" | "page"
	Number   string
	Resolved bool
}

// KindXref is the NodeKind of an Xref node.
var KindXref = gast.NewNodeKind("Xref")

// Kind implements ast.Node.Kind.
func (n *Xref) Kind() gast.NodeKind { return KindXref }

// Dump implements ast.Node.Dump.
func (n *Xref) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{"ID": n.ID, "Mode": n.Mode}, nil)
}

// NewXref returns an Xref to the given id in the given mode ("num" | "page").
func NewXref(id, mode string) *Xref { return &Xref{ID: id, Mode: mode} }
