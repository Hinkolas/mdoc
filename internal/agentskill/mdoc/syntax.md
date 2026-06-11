# mdoc body syntax

The document body is processed in this order:

1. **`:::include` splice** over the raw source.
2. **Go `text/template`** over the combined markdown body.
3. **goldmark** Markdown → HTML with GFM, footnotes, heading attributes, and
   mdoc extensions.
4. **KaTeX** renders math client-side in Chromium.

## Markdown (goldmark: GFM + footnotes)

Standard CommonMark plus GitHub extensions:

- Headings `#`…`######` with auto IDs. Non-ASCII letters are folded where
  possible, so `# Äußere Form` becomes `#aeussere-form`.
- **bold**, *italic*, ~~strikethrough~~, `inline code`.
- Lists, blockquotes (nestable), horizontal rules (`---` or `***`).
- **Tables** with column alignment:
  ```markdown
  | Left | Center | Right |
  | :--- | :----: | ----: |
  | a    | b      | c     |
  ```
- **Task lists**: `- [x] done` / `- [ ] todo`.
- **Footnotes**: `claim[^1]` … and later `[^1]: explanation`.
- **Autolinks** for bare URLs.
- **Raw HTML is allowed and passed through** — `<figure>`, `<img>`, `<div>`,
  etc. work inline (handy for figures with captions).
- **Heading attributes**: append `{#id}`, `{.unnumbered}`, `{.notoc}`,
  `{.intoc}`, or `{.numbered}` to a heading.

### Code blocks

Fenced code blocks with a language tag render as styled monospace blocks:

    ```go
    func main() {}
    ```

There is **no syntax highlighting** out of the box: the language tag becomes a
`language-…` class but nothing colors it unless a theme adds a highlighter.
Don't claim code will be highlighted.

## Math (KaTeX, client-side)

- Inline: `$ ... $` — e.g. `$T = \pi r^2$`.
- Display: `$$ ... $$` on their own lines.
- **Only dollar delimiters work.** `\( … \)` and `\[ … \]` do NOT — goldmark
  treats the backslashes as escapes and strips them before KaTeX runs.
- For a literal dollar sign in prose, escape it as `\$`.
- Inside math, `%` starts a KaTeX comment — write `\%` for a literal percent.
- Invalid math renders in KaTeX's error style instead of failing the build.

## Template interpolation in the body

The body is run through Go's `text/template` before markdown conversion, so you
can inject metadata:

- `{{.Title}}`, `{{.Author}}`, `{{.Tags}}`
- `{{.Page.Size}}`, `{{.Page.Margin}}`
- `{{.Data.<key>}}` — your custom frontmatter `data`
- `{{.System.Date}}` (e.g. `29 May 2026`), `{{.System.Time}}` (`15:04:05`),
  `{{.System.Version}}` (the mdoc version)

Example:

```markdown
*Prepared by {{.Author}} on {{.System.Date}} for project {{.Data.project}}.*
```

⚠️ Because the body is a template, literal `{{` and `}}` are interpreted. To
output literal braces, write `{{"{{"}}` and `{{"}}"}}`.

## Images and relative paths

Relative URLs resolve from the **root document's directory**. Keep assets next
to the root `.md` file and reference them relatively:

```markdown
![Diagram](assets/diagram.png)
```

## Multi-file documents (`:::include`)

Split a long document into one file per chapter and stitch them together from a
root/index file with `:::include <path>` — the LaTeX `\input` model. Each
`:::include` line (at the start of a line, not inside a code fence) is replaced
by the referenced file's body before parsing:

```markdown
---
mdoc: true
title: "My Book"
numbering:
  enabled: true
---

:::toc

:::include chapters/01-introduction.md
:::include chapters/02-methods.md
```

- **The root frontmatter owns all configuration** (theme, page, numbering,
  labels, data, references). It's the preamble; included files supply only body.
- **Included files may keep their own frontmatter** — it's parsed off and
  discarded on include, so each chapter is still openable on its own with
  `mdoc open chapters/01-introduction.md` for a focused preview.
- Numbering, the TOC, cross-references (`[#id]`) and citations all work across
  files, because the includes are spliced into one combined source first.
- Include paths resolve relative to the **including** file; includes can nest. A
  cycle or a missing file is an error.
- **Asset caveat:** the combined body renders as if it lived in the root's
  directory, so a relative `![](…)` in a chapter resolves against the *root*, not
  the chapter's folder. Put shared images under the root's tree (e.g. a top-level
  `assets/`).

## Document regions and page breaks

Single-line directives beginning with `:::` divide the document into regions or
emit generated content. Region markers are leaf directives: they do not need a
closing fence.

```markdown
:::frontmatter
# Abstract

:::toc depth=3

:::mainmatter
# Introduction

:::appendix
# Supplemental Tables
```

- `:::frontmatter` wraps following blocks in
  `<div class="mdoc-matter-front">`; headings are unnumbered and out of the TOC
  by default.
- `:::mainmatter` wraps following blocks in `<div class="mdoc-matter-main">`.
- `:::appendix` wraps following blocks in
  `<div class="mdoc-matter-appendix">`; numbered top-level headings become
  `A`, `B`, …
- Content before the first marker is left as ordinary body content.
- `:::page` emits a page break: `<div class="mdoc-pagebreak"></div>`.
- `:::page cover` adds a named style class:
  `<div class="mdoc-pagebreak mdoc-page-cover"></div>`. Themes decide how to
  use that class.

## Heading numbering and TOC

Heading numbering is off by default. Enable it in frontmatter:

```yaml
numbering:
  enabled: true
```

```markdown
:::toc depth=2

:::mainmatter
# Introduction
## Background
## Acknowledgements {.unnumbered}
## Internal Notes {.notoc}
## Executive Summary {.unnumbered .intoc}
```

- `:::toc` renders a generated table of contents.
- `depth=N` includes headings through level `N`; default is `3`.
- Numbered headings get a leading `<span class="mdoc-secnum">`.
- `{.unnumbered}` removes the visible number but keeps the heading in the TOC
  unless `.notoc` is also present.
- `{.notoc}` excludes a heading from the TOC.
- `{.intoc}` forces a heading into the TOC, useful in front matter.
- `{.numbered}` allows numbering only when `numbering.enabled` is true; it does
  not override a globally disabled document.
- `{#custom-id}` sets the anchor used by links and `[#custom-id]`.

## Figures, tables, lists, and captions

`:::figure` and `:::table` are container directives. They open with
`:::figure [#id]` / `:::table [#id]` and close with a bare `:::` line.

```markdown
:::figure #pipeline
![Pipeline overview](assets/pipeline.svg)

The ingestion pipeline after the queue split.
:::

:::table #metrics
| Metric | Value |
| :----- | ----: |
| p95    | 120ms |

Measured under normal production load.
:::
```

- Figures/tables are numbered in document order. Within numbered chapters they
  use chapter-local numbers like `2.1`; before any numbered chapter they use a
  continuous count like `1`, `2`.
- Without an explicit id, mdoc generates `fig-2-1` or `tab-2-1` from the number.
- Figure captions go below the media. Table captions go above the table.
- In a figure, image-only paragraphs are media. Other paragraphs become the
  caption. If there is no caption text, the first image alt text is used for the
  list of figures.
- Caption labels default to `Figure` / `Table`; override via frontmatter
  `labels.figure` and `labels.table`.
- `:::lof` renders a generated list of figures.
- `:::lot` renders a generated list of tables.

## Cross-references and page references

Use `[#id]` for the target number and `[#id page]` for the target page:

```markdown
As shown in Figure [#pipeline] on page [#pipeline page], latency improved.
See Section [#background].
```

- Number references link to headings, figures, and tables.
- A numberless heading reference falls back to the heading title.
- Page references render as empty links with class `mdoc-pageref`; themes fill
  the page number using paged.js `target-counter`.
- Unresolved references render as `[?]` with `mdoc-xref-unresolved`.
- `[#id](url)` and `[#id][ref]` remain ordinary markdown links, not mdoc xrefs.

## Citations and bibliography

Declare `references` in frontmatter, cite them with `[@key]`, and place the
generated list with `:::bibliography`.

```markdown
This follows the literate-programming idea [@knuth1984].
For a page locator, write [@knuth1984, p. 99].

# References

:::bibliography
```

- Citations are numbered by first appearance and render as `[1]` links.
- `[@key, locator]` parses the locator, but the current renderer displays only
  the numeric citation link.
- Missing citations render as `[?]` with `mdoc-cite-unresolved`.
- `:::bibliography` lists only cited references.
