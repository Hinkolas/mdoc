---
name: mdoc
description: >-
  Author and edit mdoc documents — markdown rendered to a paginated PDF via
  paged.js and Chromium. Use when creating or editing markdown files that carry
  `mdoc: true` frontmatter, when the user mentions mdoc, or when they want a
  PDF/preview from a markdown document using the mdoc CLI. Covers the
  frontmatter schema, supported markdown + KaTeX math, Go-template
  interpolation in the body, themes, and the `mdoc` commands.
---

# Authoring mdoc documents

mdoc renders a single markdown file (with YAML frontmatter) into a paginated
PDF. The pipeline is: **Go text/template over the body → goldmark Markdown→HTML
→ an HTML theme wrap → paged.js pagination in Chromium → PDF**.

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

- **frontmatter.md** — every frontmatter field, types, defaults, and the `page`
  size/margin and `data` map.
- **syntax.md** — supported markdown (GFM, footnotes, raw HTML), KaTeX math
  delimiters and gotchas, and Go-template interpolation in the body.
- **cli.md** — `mdoc print` / `open` / `bundle` / `install`, their flags, and
  how relative asset paths and themes resolve.
- **examples/document.md** — a complete sample document.
- **examples/plain.html** — a working starter theme to copy to
  `themes/<name>.html` and customize (the built-in `system` theme is similar).

## Producing a PDF (typical flow)

1. Write the `.md` file with `mdoc: true` frontmatter.
2. Pick a theme. Omit `theme` (or `theme: system`) for the built-in styled
   default; `theme: none` for a bare render. A custom `theme: <name>` needs
   `themes/<name>.html` next to the document or in `~/.config/mdoc/themes/` —
   copy `examples/plain.html` there as a starting point. A missing/broken
   theme isn't fatal: it falls back to `system` with a warning. (See cli.md.)
3. Run `mdoc print <file>` for a PDF, or `mdoc open <file>` for a live preview.
