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
	reg.Register(KindCitation, r.renderCitation)
	reg.Register(KindSecNum, r.renderSecNum)
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
	}
	return gast.WalkSkipChildren, nil
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
