// Package render is the single rendering pipeline shared by `mdoc print` and
// `mdoc open`. It turns a parsed document + resolved theme into the final HTML
// string that paged.js will paginate inside Chromium.
package render

import (
	"bytes"
	_ "embed"
	"fmt"
	htmltmpl "html/template"
	"time"

	texttmpl "text/template"

	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/mdext"
	"github.com/hinkolas/mdoc/internal/theme"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed shell.html
var shellTemplate string

// SystemData are values made available to templates that aren't part of the
// document's own frontmatter (timestamps, mdoc version, etc.).
type SystemData struct {
	Date    string
	Time    string
	Version string
}

// ThemeData is what theme templates and the markdown body template see.
type ThemeData struct {
	Title  string
	Author string
	Tags   []string
	Page   document.Page
	Data   map[string]any
	Body   htmltmpl.HTML
	System SystemData
}

// Options configure where browser-visible asset URLs point.
type Options struct {
	// VendorBase is the URL prefix from which paged.js + KaTeX assets load.
	// Examples: "/_/vendor" for the preview server, or a file:// URL when
	// the print pipeline extracts vendor files to a temp directory.
	VendorBase string
	// BaseHref controls how relative URLs inside the document resolve.
	// Typically a file:// URL to the document's directory for print, or
	// the server origin for preview.
	BaseHref string
	// HeadInject is extra HTML appended to <head>, e.g. live-reload glue.
	HeadInject htmltmpl.HTML
	// Version is reported as System.Version inside templates.
	Version string
}

// shellData drives shell.html. URLs are wrapped in template.URL so
// html/template doesn't refuse to emit file:// links.
type shellData struct {
	Title      string
	BaseHref   htmltmpl.URL
	VendorBase htmltmpl.URL
	HeadInject htmltmpl.HTML
	ThemedHTML htmltmpl.HTML
}

// RenderThemed runs the markdown body template, Markdown -> HTML, and the
// theme template — but NOT the shell wrap. The result is the themed HTML
// snippet that paged.js will paginate. Used by the preview server, which
// hosts its own copy of paged.js on the client side.
func RenderThemed(doc *document.Document, thm *theme.Theme, opts Options) (string, ThemeData, error) {
	td := themeData(doc, opts)

	// 1. Template pass over the markdown body so the user can interpolate
	//    metadata like `{{.Title}}` inside their markdown.
	bodyTmpl, err := texttmpl.New("body").Parse(doc.Body)
	if err != nil {
		return "", td, fmt.Errorf("parse body template: %w", err)
	}
	var mdBuf bytes.Buffer
	if err := bodyTmpl.Execute(&mdBuf, td); err != nil {
		return "", td, fmt.Errorf("execute body template: %w", err)
	}

	// 2. Markdown -> HTML. The mdext extension adds section numbering, the
	//    :::toc / :::bibliography / :::figure / :::lof directives, [@key]
	//    citations, and [#id] cross-references from the document's frontmatter.
	//    It is built per render so it sees this document's references, numbering,
	//    and caption labels.
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			mdext.New(mdext.Config{
				References: doc.Config.References,
				Numbering:  doc.Config.Numbering,
				Labels:     doc.Config.Labels,
			}),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithHeadingAttribute(), // {.unnumbered} / {.notoc} / {.appendix} / {#id}
		),
		goldmark.WithRendererOptions(html.WithUnsafe()), // themes are trusted; allow raw HTML in body
	)
	// A transliterating id generator keeps non-ASCII heading anchors stable
	// (e.g. "Äußere Form" -> "aeussere-form" instead of the lossy default).
	ctx := parser.NewContext(parser.WithIDs(mdext.NewIDs()))
	var bodyHTML bytes.Buffer
	if err := md.Convert(mdBuf.Bytes(), &bodyHTML, parser.WithContext(ctx)); err != nil {
		return "", td, fmt.Errorf("convert markdown: %w", err)
	}
	td.Body = htmltmpl.HTML(bodyHTML.String())

	// 3. Theme wrap.
	var themed bytes.Buffer
	if err := thm.Template.Execute(&themed, td); err != nil {
		return "", td, fmt.Errorf("execute theme template: %w", err)
	}
	return themed.String(), td, nil
}

// Render runs the full pipeline: RenderThemed + shell wrap. The result is a
// complete HTML document Chromium can load directly and have paged.js
// paginate. Used by the print pipeline.
func Render(doc *document.Document, thm *theme.Theme, opts Options) (string, error) {
	themed, td, err := RenderThemed(doc, thm, opts)
	if err != nil {
		return "", err
	}
	shell, err := htmltmpl.New("shell").Parse(shellTemplate)
	if err != nil {
		return "", fmt.Errorf("parse shell template: %w", err)
	}
	var out bytes.Buffer
	err = shell.Execute(&out, shellData{
		Title:      td.Title,
		BaseHref:   htmltmpl.URL(opts.BaseHref),
		VendorBase: htmltmpl.URL(opts.VendorBase),
		HeadInject: opts.HeadInject,
		ThemedHTML: htmltmpl.HTML(themed),
	})
	if err != nil {
		return "", fmt.Errorf("execute shell template: %w", err)
	}
	return out.String(), nil
}

func themeData(doc *document.Document, opts Options) ThemeData {
	now := time.Now()
	version := opts.Version
	if version == "" {
		version = "dev"
	}
	return ThemeData{
		Title:  doc.Config.Title,
		Author: doc.Config.Author,
		Tags:   doc.Config.Tags,
		Page:   doc.Config.Page,
		Data:   doc.Config.Data,
		System: SystemData{
			Date:    now.Format("02 January 2006"),
			Time:    now.Format("15:04:05"),
			Version: version,
		},
	}
}
