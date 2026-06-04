package mdext

import (
	"strconv"
	"strings"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// nodeRenderer emits the stable mdoc-* class contract for the custom nodes.
// Page numbers are deliberately not emitted: a theme adds them to TOC entries
// via paged.js `target-counter(attr(href url), page)`.
type nodeRenderer struct{}

// NewNodeRenderer returns the renderer for Directive, Citation and SecNum nodes.
func NewNodeRenderer() renderer.NodeRenderer { return &nodeRenderer{} }

// RegisterFuncs implements renderer.NodeRenderer.
func (r *nodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindDirective, r.renderDirective)
	reg.Register(KindMatter, r.renderMatter)
	reg.Register(KindCitation, r.renderCitation)
	reg.Register(KindSecNum, r.renderSecNum)
	reg.Register(KindCaptioned, r.renderCaptioned)
	reg.Register(KindCaption, r.renderCaption)
	reg.Register(KindCaptionLabel, r.renderCaptionLabel)
	reg.Register(KindXref, r.renderXref)
}

func (r *nodeRenderer) renderDirective(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkSkipChildren, nil
	}
	d := n.(*Directive)
	switch d.Name {
	case "toc":
		r.renderTOC(w, d)
	case "bibliography":
		r.renderBib(w, d)
	case "lof":
		r.renderCaptionList(w, d, "lof")
	case "lot":
		r.renderCaptionList(w, d, "lot")
	case "page":
		// A page break; the optional arg names a theme page style.
		_, _ = w.WriteString(`<div class="mdoc-pagebreak`)
		if d.Arg != "" {
			_, _ = w.WriteString(` mdoc-page-`)
			_, _ = w.Write(util.EscapeHTML([]byte(d.Arg)))
		}
		_, _ = w.WriteString("\"></div>\n")
	}
	return gast.WalkSkipChildren, nil
}

func (r *nodeRenderer) renderMatter(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<div class="mdoc-matter-`)
		_, _ = w.WriteString(n.(*Matter).Region)
		_, _ = w.WriteString("\">\n")
	} else {
		_, _ = w.WriteString("</div>\n")
	}
	return gast.WalkContinue, nil
}

func (r *nodeRenderer) renderTOC(w util.BufWriter, d *Directive) {
	_, _ = w.WriteString("<nav class=\"mdoc-toc\">\n")
	for _, h := range d.Headings {
		_, _ = w.WriteString(`<a class="mdoc-toc-entry" data-level="`)
		_, _ = w.WriteString(strconv.Itoa(h.Level))
		_, _ = w.WriteString(`" href="#`)
		_, _ = w.Write(util.EscapeHTML([]byte(h.ID)))
		_, _ = w.WriteString(`">`)
		if h.Number != "" {
			_, _ = w.WriteString(`<span class="mdoc-toc-num">`)
			_, _ = w.Write(util.EscapeHTML([]byte(h.Number)))
			_, _ = w.WriteString(`</span>`)
		}
		_, _ = w.WriteString(`<span class="mdoc-toc-text">`)
		_, _ = w.Write(util.EscapeHTML([]byte(h.Title)))
		_, _ = w.WriteString("</span></a>\n")
	}
	_, _ = w.WriteString("</nav>\n")
}

func (r *nodeRenderer) renderBib(w util.BufWriter, d *Directive) {
	_, _ = w.WriteString("<ol class=\"mdoc-bib\">\n")
	for _, e := range d.Bib {
		_, _ = w.WriteString(`<li class="mdoc-bib-entry" id="`)
		_, _ = w.WriteString(refID(e.Key))
		_, _ = w.WriteString(`"><span class="mdoc-bib-label">[`)
		_, _ = w.WriteString(strconv.Itoa(e.Number))
		_, _ = w.WriteString(`]</span><span class="mdoc-bib-text">`)
		if strings.TrimSpace(e.Ref.Text) != "" {
			_, _ = w.WriteString(e.Ref.Text) // raw escape-hatch, emitted verbatim
		} else {
			_, _ = w.Write(util.EscapeHTML([]byte(formatReference(e.Ref))))
		}
		_, _ = w.WriteString("</span></li>\n")
	}
	_, _ = w.WriteString("</ol>\n")
}

// renderCaptionList emits a `:::lof` / `:::lot` list (class "lof" or "lot").
// Page numbers are left to the theme's target-counter, like the TOC.
func (r *nodeRenderer) renderCaptionList(w util.BufWriter, d *Directive, class string) {
	_, _ = w.WriteString(`<nav class="mdoc-` + class + "\">\n")
	for _, e := range d.Entries {
		_, _ = w.WriteString(`<a class="mdoc-` + class + `-entry" href="#`)
		_, _ = w.Write(util.EscapeHTML([]byte(e.ID)))
		_, _ = w.WriteString(`">`)
		if e.Number != "" {
			_, _ = w.WriteString(`<span class="mdoc-` + class + `-num">`)
			_, _ = w.Write(util.EscapeHTML([]byte(e.Number)))
			_, _ = w.WriteString(`</span>`)
		}
		_, _ = w.WriteString(`<span class="mdoc-` + class + `-text">`)
		_, _ = w.Write(util.EscapeHTML([]byte(e.Title)))
		_, _ = w.WriteString("</span></a>\n")
	}
	_, _ = w.WriteString("</nav>\n")
}

// renderCaptioned wraps a figure/table in a <figure> the theme can style and
// number; the injected label carries the visible "Abbildung 2.1".
func (r *nodeRenderer) renderCaptioned(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	c := n.(*Captioned)
	if entering {
		class := "mdoc-figure"
		if c.Variant == "table" {
			class = "mdoc-table"
		}
		_, _ = w.WriteString(`<figure class="` + class + `" id="`)
		_, _ = w.Write(util.EscapeHTML([]byte(c.ID)))
		_, _ = w.WriteString("\">\n")
	} else {
		_, _ = w.WriteString("</figure>\n")
	}
	return gast.WalkContinue, nil
}

func (r *nodeRenderer) renderCaption(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<figcaption class="mdoc-figcaption">`)
	} else {
		_, _ = w.WriteString("</figcaption>\n")
	}
	return gast.WalkContinue, nil
}

func (r *nodeRenderer) renderCaptionLabel(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	l := n.(*CaptionLabel)
	_, _ = w.WriteString(`<span class="`)
	_, _ = w.WriteString(l.Class)
	_, _ = w.WriteString(`">`)
	_, _ = w.Write(util.EscapeHTML([]byte(l.Label)))
	_, _ = w.WriteString(`</span> `)
	return gast.WalkSkipChildren, nil
}

// renderXref emits a cross-reference: a number link (`mdoc-xref`) or an empty
// page link (`mdoc-pageref`) the theme fills via target-counter.
func (r *nodeRenderer) renderXref(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	x := n.(*Xref)
	switch {
	case !x.Resolved:
		_, _ = w.WriteString(`<span class="mdoc-xref mdoc-xref-unresolved">[?]</span>`)
	case x.Mode == "page":
		_, _ = w.WriteString(`<a class="mdoc-pageref" href="#`)
		_, _ = w.Write(util.EscapeHTML([]byte(x.ID)))
		_, _ = w.WriteString(`"></a>`)
	default:
		_, _ = w.WriteString(`<a class="mdoc-xref" href="#`)
		_, _ = w.Write(util.EscapeHTML([]byte(x.ID)))
		_, _ = w.WriteString(`">`)
		_, _ = w.Write(util.EscapeHTML([]byte(x.Number)))
		_, _ = w.WriteString(`</a>`)
	}
	return gast.WalkSkipChildren, nil
}

func (r *nodeRenderer) renderCitation(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	c := n.(*Citation)
	if c.Resolved {
		_, _ = w.WriteString(`<a class="mdoc-cite" href="#`)
		_, _ = w.WriteString(c.RefID)
		_, _ = w.WriteString(`">[`)
		_, _ = w.WriteString(strconv.Itoa(c.Number))
		_, _ = w.WriteString(`]</a>`)
	} else {
		_, _ = w.WriteString(`<span class="mdoc-cite mdoc-cite-unresolved">[?]</span>`)
	}
	return gast.WalkSkipChildren, nil
}

func (r *nodeRenderer) renderSecNum(w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<span class="mdoc-secnum">`)
	_, _ = w.Write(util.EscapeHTML([]byte(n.(*SecNum).Num)))
	_, _ = w.WriteString(`</span> `)
	return gast.WalkSkipChildren, nil
}
