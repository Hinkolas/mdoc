# mdoc

A small command-line tool that turns Markdown into print-quality PDFs. Layout is done by [paged.js](https://pagedjs.org) inside a headless Chromium, which means real CSS pagination — `@page` rules, running headers and footers, page counters, named pages, break controls — rather than the one-long-page output you get from most Markdown-to-PDF converters.

It ships with a live preview that reloads in place on every save, so you can iterate on a document without ever leaving the editor.

## Highlights

- **Real pagination** via paged.js: `@page` size and margins, margin boxes, `break-before/after/inside`, orphan/widow control, counters.
- **Live preview** in a chromeless app window with Vite-style in-place updates — no flicker, no scroll jump on save.
- **What you see is what you print**: preview and PDF go through the same render pipeline, so the on-screen pages match the PDF page-for-page.
- **GitHub Flavored Markdown** (tables, task lists, strikethrough, autolinks), [KaTeX](https://katex.org/) math (`\(...\)`, `\[...\]`, `$$...$$`), and footnotes.
- **Themes** are plain HTML + CSS with Go `html/template` placeholders. Per-document page size and margins can be overridden from the YAML frontmatter.
- **Single binary**: paged.js and KaTeX (with fonts) are embedded; the only external dependency is Chromium, which `mdoc install` will fetch for you.

## Install

```bash
curl -fsSL https://github.com/Hinkolas/mdoc/releases/latest/download/install.sh | sh
mdoc install
```

This installs the latest release binary for your platform (Linux and macOS, amd64 and arm64) into `~/.local/bin`. Set `MDOC_BIN_DIR` to install elsewhere, or `MDOC_VERSION` (e.g. `MDOC_VERSION=v0.1.0`) to pin a specific version.

`mdoc install` opens a small setup wizard in an interactive terminal. It checks for packaged/system Chromium, asks whether to download mdoc's packaged Chromium snapshot when useful, and can install the bundled mdoc authoring skill for Claude and/or Codex. In non-interactive terminals, plain `mdoc install` keeps the script-friendly behavior and downloads the packaged Chromium snapshot into `$XDG_CACHE_HOME/mdoc/chromium` (`~/Library/Caches/mdoc/chromium` on macOS). If a system Chromium is already on your `PATH`, mdoc can use it as a fallback.

## Quick start

```bash
mdoc print example/document.md     # writes example/document.pdf
mdoc open  example/document.md     # opens a live preview window
```

In the preview window, **Print** generates a PDF from the current document and downloads it; **Reload** forces a full re-render if anything ever looks stuck.

## Commands

### `mdoc print <file>`

One-shot render to PDF. Output goes next to the source as `<basename>.pdf` unless overridden.

```
-o, --output <path>   write the PDF here instead
    --html            also write the rendered HTML next to the PDF (debugging)
```

### `mdoc open <file>`

Opens a chromeless Chromium window with the document. The server watches the document and its resolved theme for changes; on save it pushes a reload signal and the iframe re-paginates in place (scroll position preserved).

```
-p, --port <n>   preview server port (default 7768, 0 picks a free one)
```

### `mdoc install`

Runs the interactive setup wizard for Chromium and optional agent skills. In non-interactive terminals, downloads Chromium into the user cache directory.

```
    --chromium[=<revision>]   install packaged Chromium, optionally pinned to a revision
    --skill <target>          install bundled skill: claude, codex, or all
    --path <dir>              parent skills dir for a single --skill target
```

Examples:

```bash
mdoc install --chromium
mdoc install --chromium=123456
mdoc install --skill claude
mdoc install --skill codex
mdoc install --skill all
mdoc install --skill claude --path ~/agent/skills
```

## Document format

Each document is a Markdown file with an optional YAML frontmatter block. The `mdoc: true` field opts the file into the rendering pipeline — without it, defaults are used.

```markdown
---
mdoc: true
theme: plain
title: "Quarterly Report"
author: "Jane Doe"
tags: [report, 2026]
page:
  size: Letter
  margin: 1in
data:
  client: "Acme Corp"
---

# {{.Title}}

Prepared for {{.Data.client}} by {{.Author}}.
```

### Frontmatter fields

| Field          | Purpose                                                              |
| -------------- | -------------------------------------------------------------------- |
| `mdoc`         | Set to `true` to enable mdoc rendering for this file.                |
| `theme`        | Theme name; resolved against `./themes/`, then `~/.config/mdoc/themes/`, then a built-in (`system`/`none`). Defaults to `system`. |
| `title`        | Document title; exposed as `{{.Title}}`.                             |
| `author`       | Author name; exposed as `{{.Author}}`.                               |
| `tags`         | List of tags; exposed as `{{.Tags}}`.                                |
| `page.size`    | CSS `@page` size: `A4`, `Letter`, `A4 landscape`, `210mm 297mm`, ... |
| `page.margin`  | CSS `@page` margin: `25mm`, `1in`, `25mm 22mm 28mm 22mm`, ...        |
| `data`         | Arbitrary map exposed as `{{.Data.<key>}}`.                          |

System values like `{{.System.Date}}`, `{{.System.Time}}`, and `{{.System.Version}}` are available in both the Markdown body and the theme template.

`page.size` and `page.margin` are passed through verbatim into the theme's `@page` rule, so anything CSS accepts works — the theme decides what its fallback is when you leave them empty.

### Math, code, tables

- `$ ... $` — inline math (escape a literal dollar with `\$`)
- `$$ ... $$` — display math
- Triple-backtick fences for code, including ` ```diff ` for diff blocks
- Pipe tables with column alignment (`:---`, `:---:`, `---:`)
- `- [ ]` / `- [x]` task lists

See `example/document.md` for a doc that exercises all of these.

## Themes

A theme is an HTML file that wraps the rendered Markdown body. The file is processed by Go's `html/template`, so you can interpolate any of the document fields.

Themes are resolved in this order:

1. `<document_dir>/themes/<name>.html`
2. `~/.config/mdoc/themes/<name>.html` (override the base with `$XDG_CONFIG_HOME`)
3. a built-in keyword: **`system`** (the styled default, used when `theme` is omitted) or **`none`** (bare rendered body, no styling)

A theme file on disk overrides a built-in of the same name, so dropping a `themes/system.html` restyles every document that doesn't name a theme. A `theme:` that can't be found or won't parse falls back to `system` with a warning rather than failing.

A minimal theme:

```html
<style>
    @page {
        size: {{or .Page.Size "A4"}};
        margin: {{or .Page.Margin "25mm"}};
        @bottom-center { content: counter(page); }
    }
    body { font-family: Georgia, serif; font-size: 11pt; }
</style>

{{.Body}}
```

The `{{or .Page.Size "A4"}}` pattern lets the document override the page size from its frontmatter, falling back to the theme's default when it doesn't.

A more complete reference theme lives at `example/themes/plain.html` — serif body, sans-serif headings, page numbers in the bottom margin (suppressed on the first page), and sensible defaults for tables, code, blockquotes, and task lists.

## Generated content: contents, numbering, figures, bibliography

mdoc builds a table of contents, section numbers, figures and tables with their lists, and a bibliography from the markdown itself — you don't hand-write them. Single-line **directives** (`:::name [options]`) mark where generated blocks go and divide the document into regions; container directives (`:::figure … :::`) wrap a body. `[@key]` cites references declared in frontmatter and `[#id]` cross-references a numbered element.

### Document regions

Three markers switch the current *matter*; everything up to the next marker belongs to it. Regions set the numbering and TOC defaults, so you rarely mark individual headings:

```markdown
:::frontmatter      ← headings unnumbered and out of the TOC (abstract, …)
# Kurzreferat

:::mainmatter       ← numbered chapters, in the TOC; each starts a new page
# Introduction

:::appendix         ← top-level headings become A, B, … (in the TOC)
# Diagrams
```

`:::page` forces a page break where the engine wouldn't — e.g. between two front-matter pages; `:::page <style>` names a theme page style. Chapters in `:::mainmatter` / `:::appendix` break automatically, so `:::page` is mostly a front-matter tool.

### Multi-file documents

A long document can be split into one file per chapter and stitched back together from a root/index file with `:::include`, the way LaTeX uses `\input`. The root owns the layout and all configuration; each `:::include <path>` is replaced by the referenced file's body, in place:

```markdown
:::frontmatter
:::toc

:::mainmatter
:::include chapters/01-introduction.md
:::include chapters/02-methods.md

:::appendix
:::include appendix/data.md
```

Includes are spliced **before** the document is parsed, so everything that spans chapters just works from one combined source: continuous heading numbering, a TOC over every chapter, per-chapter figure/table counters, a single bibliography, and a `[#id]` cross-reference that points from one chapter into another.

- **The root owns configuration.** Theme, `page`, `numbering`, `labels`, `data`, and `references` all come from the root frontmatter — the LaTeX preamble model.
- **Chapters may keep their own frontmatter.** It's parsed off and discarded on include, so a chapter file stays independently openable with `mdoc open chapters/01-introduction.md` for focused editing while carrying its own `mdoc: true` / `theme:` for that standalone preview.
- **Paths resolve relative to the including file**, so a `part1/index.md` can `:::include chapter.md` from its own directory. Includes nest; a cycle or a missing file is a clear error.
- **Relative asset paths resolve relative to the *root* document.** The combined body renders as if it all lived in the root's directory, so shared images belong under the root's tree (e.g. a top-level `assets/`). Per-chapter asset directories are a known limitation — see `example/thesis/LIMITATIONS.md`.
- `mdoc open` watches every included file, so editing a chapter live-reloads the preview; `mdoc bundle` packs all of them into the `.mdoc` archive at their relative paths.

`example/book/` is a small worked example: a root with a preface and a generated TOC, two chapters in `chapters/`, and a cross-reference running between them.

### Table of contents

Put a `:::toc` where the contents should appear; mdoc collects every heading in document order and renders the list. Page numbers are filled in at print time, so they stay correct.

```markdown
# Contents

:::toc depth=3
```

- `depth=N` — deepest heading level to include (default `3`).

### Heading numbering

Section numbers (`1`, `1.1`, `A.1`) are **off by default** so ordinary documents don't get number prefixes. Turn them on in frontmatter; they then appear in the headings *and* the TOC from one source:

```yaml
numbering:
  enabled: true
```

Per-heading markers override the region default (written as a trailing `{…}`):

| Marker | Effect |
| --- | --- |
| `## Title {.unnumbered}` | drop the section number (stays in the TOC) |
| `## Title {.notoc}` | exclude from the TOC |
| `## Title {#my-id}` | explicit anchor id (otherwise auto-slugged, with `ä→ae`, `ß→ss`, …) |

### Citations and bibliography

Declare references in frontmatter, cite them inline with `[@key]`, and place the list with `:::bibliography`. Citations are numbered by first appearance; only cited references are listed.

```yaml
references:
  - key: lanze1982
    author: "Lanze, Werner"
    title: "Das technische Manuskript"
    year: "1982"
    publisher: "Vulkan-Verlag"
  - key: site
    text: "Raw, hand-formatted entry, emitted verbatim (may contain HTML)."
```

```markdown
… a point made by Lanze [@lanze1982] …

# References

:::bibliography
```

Each reference takes `author`, `title`, `year`, `publisher`, `edition`, `isbn`, `url`, or a raw `text:` escape-hatch used verbatim. Richer citation styles (CSL) are future work.

### Figures and tables

`:::figure` and `:::table` are container directives: their markdown body holds the media (a real `![alt](src)` image, or a markdown table) and a **rich caption** — the caption is ordinary markdown, so bold, italics, links, `[@key]` citations and `[#id]` cross-references all work inside it. mdoc numbers them per chapter (`2.1`, `A.1` in the appendix), injects the `Figure 2.1` / `Table 2.1` label, and lists them from `:::lof` / `:::lot`.

```markdown
:::figure #fig-voltage
![Voltage trace](assets/plot.svg)

Harmonic voltage at **50 Hz**, see [#sec-method].
:::

:::table #tab-margins
| Edge | Margin |
| :--- | :----: |
| top  | 2.5    |

Required page margins.
:::
```

- Image-only paragraphs are the media; the remaining text is the caption. Several images in one paragraph become side-by-side subfigures.
- A figure's caption sits below the media; a table's caption sits above it.
- `#id` is optional (one is generated for unlabelled figures); use it to cross-reference the figure.
- Place the lists with `:::lof` (figures) and `:::lot` (tables); both fill page numbers at print time, like the TOC.

The caption word is `Figure` / `Table` by default; override per language in frontmatter:

```yaml
labels:
  figure: "Abbildung"
  table: "Tabelle"
```

### Cross-references

`[#id]` prints the number of the heading, figure or table with that id and links to it; `[#id page]` prints its page number instead (resolved by the theme at print time). You supply the surrounding noun, so it reads naturally in any language:

```markdown
… see Figure [#fig-voltage] in Section [#sec-method] on page [#sec-method page].
```

An id that resolves to no element renders as `[?]`. Headings are referenced by their auto-slug (or an explicit `{#id}`); figures/tables by their `#id`.

### Theme CSS classes

Generated blocks emit a stable, `mdoc-`-prefixed class contract for themes to style. Page numbers are **not** emitted — a theme adds them via paged.js `target-counter`.

| Block | HTML structure |
| --- | --- |
| TOC | `<nav class="mdoc-toc">` › `<a class="mdoc-toc-entry" data-level="N" href="#id">` › `<span class="mdoc-toc-num">` + `<span class="mdoc-toc-text">` |
| Section number | `<span class="mdoc-secnum">2.1</span>` as the heading's first child |
| Citation | `<a class="mdoc-cite" href="#mdoc-ref-KEY">[1]</a>` — unresolved: `<span class="mdoc-cite mdoc-cite-unresolved">[?]</span>` |
| Bibliography | `<ol class="mdoc-bib">` › `<li class="mdoc-bib-entry" id="mdoc-ref-KEY">` › `<span class="mdoc-bib-label">[1]</span>` + `<span class="mdoc-bib-text">` |
| Figure / table | `<figure class="mdoc-figure">` / `mdoc-table` › media + `<figcaption class="mdoc-figcaption">` › `<span class="mdoc-fig-label">` / `mdoc-tab-label` + caption |
| List of figures/tables | `<nav class="mdoc-lof">` / `mdoc-lot` › `<a class="mdoc-lof-entry" href="#id">` › `<span class="mdoc-lof-num">` + `<span class="mdoc-lof-text">` |
| Cross-reference | `<a class="mdoc-xref" href="#id">2.1</a>` — page: `<a class="mdoc-pageref" href="#id"></a>` — unresolved: `<span class="mdoc-xref mdoc-xref-unresolved">[?]</span>` |
| Matter region | `<div class="mdoc-matter-front">` / `-main` / `-appendix` wrapping the region |
| Page break | `<div class="mdoc-pagebreak"></div>` (optionally `mdoc-page-<style>`) |

Page numbers are filled in by the theme via `target-counter` — for the TOC, the lists of figures/tables, and `[#id page]` references:

```css
.mdoc-toc-entry::after,
.mdoc-lof-entry::after,
.mdoc-lot-entry::after { content: target-counter(attr(href), page); }
.mdoc-pageref::after  { content: target-counter(attr(href), page); }
```

`example/thesis/` is a full worked example: cover page, numbered chapters, a generated TOC, lists of figures and tables, `:::figure` / `:::table` with rich captions, `[@key]` citations with a bibliography, `[#id]` cross-references, and lettered appendices — with no hand-written apparatus in the body.

## How it works

```
foo.md  ──▶  :::include splice  ──▶  Goldmark (GFM + footnotes)  ──▶  HTML body
                                              │
                                              ▼
                              html/template (theme)
                                              │
                                              ▼
                          shell.html (paged.js + KaTeX)
                                              │
              ┌────────────────── HTTP ──────┴──────┐
              ▼                                     ▼
       headless Chromium                  preview iframe
       (mdoc print)                       (mdoc open)
              │                                     │
              ▼                                     ▼
        page.PDF()                       PagedModule.Previewer
              │                                     │
              ▼                                     ▼
         document.pdf                     pages in your window
```

Both pipelines share the same render code, so the PDF and the preview are produced from the same HTML. The preview re-paginates in place using a double-buffered swap inside the iframe — the new pages are built in a hidden sibling element and atomically swapped in, which is why edits don't flash or jump.

## Roadmap

Rough list of things on the horizon. Open to ideas — file an issue if any of these would matter to you.

- **First-class figure syntax.** Today figures (image + caption + attribution) need raw `<figure>` / `<figcaption>` HTML, which is verbose and out of place in a Markdown document. A shorthand like `![alt](path "caption")` extending into a real figure, or a fenced block syntax, would make this a one-liner.
- **Auto-generated figure index.** Once figures are first-class, a `{{.Figures}}` table-of-figures (numbering, captions, page references) the theme can render somewhere — same idea as a table of contents.
- **`mdoc import` to round-trip `.mdoc` bundles** back into a working directory.
- **More built-in themes** beyond `plain` — at least a contract/letter style and an article style.

## License

mdoc is licensed under the MIT License — see `LICENSE`.

mdoc embeds the following third-party assets, each retaining its own license:

- [paged.js](https://pagedjs.org), [MIT license](https://gitlab.coko.foundation/pagedjs/pagedjs/-/blob/main/LICENSE.md)
- [KaTeX](https://katex.org), [MIT license](https://github.com/KaTeX/KaTeX?tab=MIT-1-ov-file#readme)
- [KaTeX fonts](https://github.com/KaTeX/katex-fonts), [MIT license](https://github.com/KaTeX/katex-fonts/blob/master/LICENSE)

For the Go libraries mdoc builds on, run `go list -m all` or see `go.mod`.
