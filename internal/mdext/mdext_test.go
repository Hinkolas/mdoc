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
	got := render(t, mdext.Config{}, ":::toc\n\n# Eins\n## Zwei\n")
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
		"",
		"# Kurzreferat {.unnumbered .notoc}",
		"# Einleitung",
		"# Literaturverzeichnis {.unnumbered}",
	}, "\n"))

	// {.unnumbered .notoc}: no number, excluded from the TOC.
	notAny(t, got, `href="#kurzreferat"`)
	wantAll(t, got,
		`<h1 class="unnumbered notoc" id="kurzreferat">Kurzreferat</h1>`, // no secnum span
		`<h1 id="einleitung"><span class="mdoc-secnum">1</span>`,
		// {.unnumbered} alone: no number, but still in the TOC (no num span).
		`<h1 class="unnumbered" id="literaturverzeichnis">Literaturverzeichnis</h1>`,
		`href="#literaturverzeichnis"><span class="mdoc-toc-text">Literaturverzeichnis</span>`,
	)
	notAny(t, got, `mdoc-secnum">2`) // Literaturverzeichnis didn't consume a number
}

func TestMatterRegions(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		":::frontmatter",
		"# Kurzreferat",
		"",
		":::mainmatter",
		"# Einleitung",
		"## Aufbau",
		"",
		":::appendix",
		"# Diagramme",
		"## Detail",
		"# Software",
	}, "\n"))

	wantAll(t, got,
		// regions wrap their content:
		`<div class="mdoc-matter-front">`,
		`<div class="mdoc-matter-main">`,
		`<div class="mdoc-matter-appendix">`,
		// front matter: unnumbered (no secnum):
		`<h1 id="kurzreferat">Kurzreferat</h1>`,
		// main matter: decimal:
		`<h1 id="einleitung"><span class="mdoc-secnum">1</span>`,
		`<h2 id="aufbau"><span class="mdoc-secnum">1.1</span>`,
		// appendix: lettered, with the chapter counter reset:
		`<h1 id="diagramme"><span class="mdoc-secnum">A</span>`,
		`<h2 id="detail"><span class="mdoc-secnum">A.1</span>`,
		`<h1 id="software"><span class="mdoc-secnum">B</span>`,
	)
	// markers are consumed, not rendered as empty directives:
	notAny(t, got, `:::frontmatter`, `mdoc-matter-front">\n</div>`)
}

func TestPageBreak(t *testing.T) {
	got := render(t, mdext.Config{}, "A\n\n:::page\n\nB\n")
	wantAll(t, got, `<div class="mdoc-pagebreak"></div>`)

	styled := render(t, mdext.Config{}, ":::page cover\n")
	wantAll(t, styled, `<div class="mdoc-pagebreak mdoc-page-cover"></div>`)
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
		":::toc depth=1",
		"",
		"# One",
		"## Two",
	}, "\n"))
	wantAll(t, got, `href="#one"`)
	notAny(t, got, `href="#two"`) // depth 1 excludes the h2
}

func TestFigure(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #fig-a",
		"![Alt text](img.svg)",
		"",
		"Eine *reiche* Bildunterschrift.",
		":::",
	}, "\n"))

	wantAll(t, got,
		`<figure class="mdoc-figure" id="fig-a">`,
		`<img src="img.svg" alt="Alt text">`, // media stays as a real image
		`<figcaption class="mdoc-figcaption">`,
		`<span class="mdoc-fig-label">Figure 1.1</span> `, // chapter-scoped number
		`<em>reiche</em>`,                                 // caption keeps rich inline markup
	)
	// media renders before the caption for figures.
	if strings.Index(got, "<img") > strings.Index(got, "<figcaption") {
		t.Errorf("figure caption should follow the media:\n%s", got)
	}
}

func TestFigureContinuousWithoutNumbering(t *testing.T) {
	// No chapter number (numbering off) -> figures count continuously.
	got := render(t, mdext.Config{}, strings.Join([]string{
		":::figure #one",
		"![](a.svg)",
		"",
		"First.",
		":::",
		"",
		":::figure #two",
		"![](b.svg)",
		"",
		"Second.",
		":::",
	}, "\n"))
	wantAll(t, got,
		`<span class="mdoc-fig-label">Figure 1</span>`,
		`<span class="mdoc-fig-label">Figure 2</span>`,
	)
}

func TestTable(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::table #tab-a",
		"| A | B |",
		"| --- | --- |",
		"| 1 | 2 |",
		"",
		"Eine Tabellenunterschrift.",
		":::",
	}, "\n"))

	wantAll(t, got,
		`<figure class="mdoc-table" id="tab-a">`,
		`<span class="mdoc-tab-label">Table 1.1</span> `,
		"<table>", // the markdown table survives as media
	)
	// the caption sits above the table for tables.
	if strings.Index(got, "<figcaption") > strings.Index(got, "<table>") {
		t.Errorf("table caption should precede the table:\n%s", got)
	}
}

func TestFigureAndTableCountersAreIndependent(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #f1",
		"![](a.svg)",
		"",
		"Fig one.",
		":::",
		"",
		":::table #t1",
		"| A |",
		"| --- |",
		"| 1 |",
		"",
		"Tab one.",
		":::",
		"",
		":::figure #f2",
		"![](b.svg)",
		"",
		"Fig two.",
		":::",
	}, "\n"))
	wantAll(t, got,
		`<span class="mdoc-fig-label">Figure 1.1</span>`,
		`<span class="mdoc-tab-label">Table 1.1</span>`,
		`<span class="mdoc-fig-label">Figure 1.2</span>`,
	)
}

func TestLOFAndLOT(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #fig-a",
		"![](a.svg)",
		"",
		"Cap A.",
		":::",
		"",
		":::table #tab-a",
		"| A |",
		"| --- |",
		"| 1 |",
		"",
		"Tab A.",
		":::",
		"",
		"# Verzeichnisse {.unnumbered}",
		"",
		":::lof",
		"",
		":::lot",
	}, "\n"))

	wantAll(t, got,
		`<nav class="mdoc-lof">`,
		`<a class="mdoc-lof-entry" href="#fig-a"><span class="mdoc-lof-num">1.1</span><span class="mdoc-lof-text">Cap A.</span></a>`,
		`<nav class="mdoc-lot">`,
		`<a class="mdoc-lot-entry" href="#tab-a"><span class="mdoc-lot-num">1.1</span><span class="mdoc-lot-text">Tab A.</span></a>`,
	)
}

func TestFigureCaptionlessAltFallback(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #fig-a",
		"![A lone image](a.svg)",
		":::",
		"",
		":::lof",
	}, "\n"))
	// No caption text -> the list entry falls back to the image alt.
	wantAll(t, got, `<span class="mdoc-lof-text">A lone image</span>`)
}

func TestSubfigures(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #fig-pair",
		"![links](a.svg) ![rechts](b.svg)",
		"",
		"Zwei Unterabbildungen.",
		":::",
	}, "\n"))
	// Both images are media; the text is the caption.
	wantAll(t, got,
		`<img src="a.svg" alt="links">`,
		`<img src="b.svg" alt="rechts">`,
		`<span class="mdoc-fig-label">Figure 1.1</span> Zwei Unterabbildungen.`,
	)
}

func TestFigureAppendixScoped(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		":::appendix",
		"# Diagramme",
		"",
		":::figure #fig-x",
		"![](a.svg)",
		"",
		"Im Anhang.",
		":::",
	}, "\n"))
	wantAll(t, got, `<span class="mdoc-fig-label">Figure A.1</span>`)
}

func TestCustomLabels(t *testing.T) {
	cfg := numbered()
	cfg.Labels = map[string]string{"figure": "Abbildung", "table": "Tabelle"}
	got := render(t, cfg, strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #fig-a",
		"![](a.svg)",
		"",
		"Bild.",
		":::",
	}, "\n"))
	wantAll(t, got, `<span class="mdoc-fig-label">Abbildung 1.1</span>`)
}

func TestCrossRefNumber(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"## Aufbau",
		"",
		":::figure #fig-a",
		"![](a.svg)",
		"",
		"Bild.",
		":::",
		"",
		"Siehe Abschnitt [#aufbau] und Abbildung [#fig-a].",
	}, "\n"))
	wantAll(t, got,
		`<a class="mdoc-xref" href="#aufbau">1.1</a>`,
		`<a class="mdoc-xref" href="#fig-a">1.1</a>`,
	)
}

func TestCrossRefPage(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"## Aufbau",
		"",
		"Auf Seite [#aufbau page].",
	}, "\n"))
	// Page references emit an empty link the theme fills via target-counter.
	wantAll(t, got, `<a class="mdoc-pageref" href="#aufbau"></a>`)
}

func TestCrossRefUnresolved(t *testing.T) {
	got := render(t, numbered(), "# Kapitel\n\nSee [#nope] and [#missing page].\n")
	if strings.Count(got, `mdoc-xref-unresolved`) != 2 {
		t.Errorf("expected two unresolved cross-references:\n%s", got)
	}
}

func TestCrossRefDoesNotEatLinks(t *testing.T) {
	got := render(t, mdext.Config{}, "A [#fig-a](https://example.com) link.\n")
	wantAll(t, got, `<a href="https://example.com">#fig-a</a>`)
	notAny(t, got, `mdoc-xref`)
}

func TestCaptionEntitiesDecodedInList(t *testing.T) {
	got := render(t, numbered(), strings.Join([]string{
		"# Kapitel",
		"",
		":::figure #fig-a",
		"![](a.svg)",
		"",
		"Spannung mit 50&nbsp;Hz.",
		":::",
		"",
		":::lof",
	}, "\n"))
	// The list entry decodes the entity instead of showing a literal "&nbsp;".
	wantAll(t, got, "<span class=\"mdoc-lof-text\">Spannung mit 50 Hz.</span>")
	notAny(t, got, `&amp;nbsp;`)
}
