package mdext_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/mdext"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// render mirrors internal/render's goldmark setup so tests exercise the real
// pipeline (extensions, heading attributes, transliterating ids).
func render(t *testing.T, cfg mdext.Config, md string) string {
	t.Helper()
	g := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote, mdext.New(cfg)),
		goldmark.WithParserOptions(parser.WithAutoHeadingID(), parser.WithHeadingAttribute()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	ctx := parser.NewContext(parser.WithIDs(mdext.NewIDs()))
	var buf bytes.Buffer
	if err := g.Convert([]byte(md), &buf, parser.WithContext(ctx)); err != nil {
		t.Fatalf("convert: %v", err)
	}
	return buf.String()
}

// numbered returns a Config with heading numbering switched on.
func numbered() mdext.Config {
	return mdext.Config{Numbering: document.Numbering{Enabled: true}}
}

func wantAll(t *testing.T, got string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(got, s) {
			t.Errorf("missing %q in:\n%s", s, got)
		}
	}
}

func notAny(t *testing.T, got string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if strings.Contains(got, s) {
			t.Errorf("unexpected %q in:\n%s", s, got)
		}
	}
}

func TestTOCStructureAndNumbering(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		":::toc",
		":::",
		"",
		"# Einleitung",
		"## Aufbau",
		"## Inhalt",
		"# Schluss",
	}, "\n"))

	wantAll(t, got,
		`<nav class="mdoc-toc">`,
		`<a class="mdoc-toc-entry" data-level="1" href="#einleitung">`,
		`<span class="mdoc-toc-num">1</span>`,
		`<span class="mdoc-toc-text">Einleitung</span>`,
		`data-level="2" href="#aufbau"`,
		`<span class="mdoc-toc-num">2</span><span class="mdoc-toc-text">Schluss</span>`,
		// numbers injected into the headings themselves:
		`<h1 id="einleitung"><span class="mdoc-secnum">1</span>`,
		`<h2 id="aufbau"><span class="mdoc-secnum">1.1</span>`,
		`<h2 id="inhalt"><span class="mdoc-secnum">1.2</span>`,
	)
}

func TestNumberingOffByDefault(t *testing.T) {
	got := render(t, mdext.Config{}, ":::toc\n:::\n\n# Eins\n## Zwei\n")
	// Headings carry no number, and the TOC entries have no number span,
	// but the TOC is still generated.
	wantAll(t, got,
		`<nav class="mdoc-toc">`,
		`<a class="mdoc-toc-entry" data-level="1" href="#eins"><span class="mdoc-toc-text">Eins</span></a>`,
		`<h1 id="eins">Eins</h1>`,
	)
	notAny(t, got, `mdoc-secnum`, `mdoc-toc-num`)
}

func TestHeadingMarkers(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		":::toc",
		":::",
		"",
		"# Kurzreferat {.unnumbered .notoc}",
		"# Einleitung",
		"# Anhang {.appendix}",
		"## Detail",
		"# Weiteres",
	}, "\n"))

	// Kurzreferat: unnumbered + excluded from the TOC. The marker classes
	// deliberately leak onto the heading (e.g. class="appendix").
	notAny(t, got, `href="#kurzreferat"`)
	wantAll(t, got,
		`<h1 class="unnumbered notoc" id="kurzreferat">Kurzreferat</h1>`, // no secnum span
		`<h1 id="einleitung"><span class="mdoc-secnum">1</span>`,
		// appendix lettering:
		`<h1 class="appendix" id="anhang"><span class="mdoc-secnum">A</span>`,
		`<h2 id="detail"><span class="mdoc-secnum">A.1</span>`,
		`<h1 id="weiteres"><span class="mdoc-secnum">B</span>`,
		`<span class="mdoc-toc-num">A</span><span class="mdoc-toc-text">Anhang</span>`,
	)
}

func TestCitationsAndBibliography(t *testing.T) {
	cfg := mdext.Config{References: []document.Reference{
		{Key: "smith2020", Author: "Smith, J.", Title: "A Title", Year: "2020", Publisher: "ACME"},
		{Key: "jones2019", Text: "Jones, R.: <em>Raw</em> entry."},
	}}
	got := render(t, cfg, strings.Join([]string{
		"First [@smith2020] then [@jones2019] then [@smith2020] again.",
		"An undefined one [@nope].",
		"",
		":::bibliography",
		":::",
	}, "\n"))

	wantAll(t, got,
		`<a class="mdoc-cite" href="#mdoc-ref-smith2020">[1]</a>`,
		`<a class="mdoc-cite" href="#mdoc-ref-jones2019">[2]</a>`,
		`<span class="mdoc-cite mdoc-cite-unresolved">[?]</span>`,
		`<ol class="mdoc-bib">`,
		`<li class="mdoc-bib-entry" id="mdoc-ref-smith2020"><span class="mdoc-bib-label">[1]</span>`,
		`Smith, J. (2020). A Title. ACME.`,
		// raw text escape-hatch emitted verbatim (the <em> survives):
		`<span class="mdoc-bib-text">Jones, R.: <em>Raw</em> entry.</span>`,
	)
	// second [@smith2020] reuses [1], and the undefined key is not in the bib.
	if strings.Count(got, `href="#mdoc-ref-smith2020">[1]`) != 2 {
		t.Errorf("expected smith2020 cited twice as [1]:\n%s", got)
	}
	notAny(t, got, "mdoc-ref-nope")
}

func TestCoexistsWithLinksAndFootnotes(t *testing.T) {
	cfg := mdext.Config{References: []document.Reference{{Key: "k", Text: "Entry."}}}
	got := render(t, cfg, strings.Join([]string{
		"A [link](https://example.com), a footnote[^n], and a cite [@k].",
		"",
		"[^n]: the note.",
	}, "\n"))

	wantAll(t, got,
		`<a href="https://example.com">link</a>`,
		`<a class="mdoc-cite" href="#mdoc-ref-k">[1]</a>`,
		`role="doc-noteref"`, // footnote link still rendered by goldmark
	)
}

func TestTransliteratedHeadingID(t *testing.T) {
	got := render(t, mdext.Config{}, "# Äußere Form\n")
	wantAll(t, got, `id="aeussere-form"`)
	notAny(t, got, `id="uere-form"`)
}

func TestTOCDepthOption(t *testing.T) {
	got := render(t, mdext.Config{}, strings.Join([]string{
		":::toc",
		"depth: 1",
		":::",
		"",
		"# One",
		"## Two",
	}, "\n"))
	wantAll(t, got, `href="#one"`)
	notAny(t, got, `href="#two"`) // depth 1 excludes the h2
}
