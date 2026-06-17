---
name: mdoc
description: >-
  Author and edit mdoc documents — markdown rendered to a paginated PDF via
  paged.js and Chromium. Use when creating or editing markdown files that carry
  `mdoc: true` frontmatter, when the user mentions mdoc, or when they want a
  PDF/preview/bundle from markdown using the mdoc CLI. Covers frontmatter,
  supported markdown + KaTeX math, Go-template interpolation, directives
  (`:::toc`, `:::figure`, `:::table`, `:::include`, etc.), citations,
  cross-references, themes, and the `mdoc` commands.
---

# Authoring mdoc documents

mdoc renders markdown with YAML frontmatter into a paginated PDF. The pipeline
is: **`:::include` splice → Go text/template over the body → goldmark
Markdown→HTML with mdoc extensions → HTML theme wrap → paged.js pagination in
Chromium → PDF**.

## The one rule that trips people up

A file is only treated as an mdoc document if its frontmatter opts in with
`mdoc: true`. **If `mdoc: true` is missing, every other frontmatter field is
ignored** and built-in defaults are used (the built-in `system` theme, title
"Untitled", author "Anonymous"). Always start an mdoc file with `mdoc: true`.

## Minimal document

```markdown
---
mdoc: true
title: "My Document"
author: "Jane Doe"
---

# My Document

Body in **GitHub-flavored markdown**. Inline math like $E = mc^2$ and display
math:

$$
\int_0^\infty e^{-x}\,dx = 1
$$
```

## Reference files (open the relevant one when you need detail)

- **frontmatter.md** — every frontmatter field, types, defaults, references,
  numbering, labels, page size/margin, and custom `data`.
- **syntax.md** — markdown, math, templates, directives, includes, numbering,
  figures/tables, TOC/LOF/LOT, citations, bibliography, cross-references, and
  page breaks.
- **themes.md** — theme lookup, template data, paged.js page rules, and the
  stable `mdoc-*` CSS class contract emitted by the renderer.
- **cli.md** — `mdoc print` / `open` / `bundle` / `install`, their flags, and
  script-friendly output behavior.
- **examples/document.md** — a complete sample document.
- **examples/plain.html** — a working starter theme to copy next to a document
  (`themes/plain.html`, used via `theme: ./themes/plain.html`) or into
  `~/.config/mdoc/themes/` (used via `theme: plain`) and customize (the built-in
  `system` theme is similar).

## Producing a PDF (typical flow)

1. Write the `.md` file with `mdoc: true` frontmatter.
2. Pick a theme. Omit `theme` (or `theme: system`) for the built-in styled
   default; `theme: none` for a bare render. For a custom theme, either drop a
   file in `~/.config/mdoc/themes/` and name it by key (`theme: mytheme`), or
   point at a file next to the document by path (`theme: ./themes/mytheme.html`)
   — copy `examples/plain.html` as a starting point. A missing/broken theme
   falls back to `system` with a warning. (See themes.md.)
3. Run `mdoc print <file>` for a PDF, or `mdoc open <file>` for a live preview.

## Authoring rules of thumb

- Start every real document with `mdoc: true`; otherwise frontmatter is ignored
  and defaults are used.
- Use mdoc directives instead of hand-written apparatus: `:::toc`,
  `:::figure`, `:::table`, `:::lof`, `:::lot`, `:::bibliography`, and
  `:::page`.
- Use `{#id}`, `{.unnumbered}`, `{.notoc}`, `{.intoc}` and frontmatter
  `numbering.enabled` for heading numbering/TOC control.
- Use `[#id]` for number references, `[#id page]` for page references, and
  `[@key]` for bibliography citations.
- When creating custom themes, style the stable `mdoc-*` classes documented in
  themes.md and keep `{{.Body}}` in the template.
