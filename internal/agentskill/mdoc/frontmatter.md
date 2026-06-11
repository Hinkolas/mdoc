# mdoc frontmatter

YAML at the very top of the file, between `---` fences. Parsed into the schema
below; unknown keys are ignored.

| Key | Type | Default | Notes |
|-----|------|---------|-------|
| `mdoc` | bool | — | **Required.** Must be `true`, or the entire frontmatter is discarded and defaults apply. |
| `theme` | string | `system` | Theme name → `./themes/<name>.html`, then `~/.config/mdoc/themes/<name>.html`, then a built-in. Two built-in keywords: **`system`** (a styled, dependable allrounder — the default when omitted/empty) and **`none`** (bare rendered body, no styling). A same-named theme file on disk overrides a built-in. A name that can't be found or parsed falls back to `system` with a warning — never a hard failure. |
| `title` | string | `Untitled` | HTML `<title>`; also available as `{{.Title}}`. |
| `author` | string | `Anonymous` | Available as `{{.Author}}`. |
| `tags` | string list | `[]` | Available as `{{.Tags}}`. |
| `page.size` | string | theme decides | Passed verbatim into the theme's `@page { size: … }`. CSS page-size syntax: `A4`, `Letter`, `A4 landscape`, `210mm 297mm`, … |
| `page.margin` | string | theme decides | Passed verbatim into `@page { margin: … }`. CSS margin shorthand: `25mm`, `25mm 22mm`, `25mm 22mm 28mm 22mm`. |
| `data` | map | `{}` | Arbitrary key/values, available in the body and theme as `{{.Data.<key>}}`. |
| `numbering.enabled` | bool | `false` | Enables automatic heading numbers (`1`, `1.1`, `A.1`) and numbered TOC entries. |
| `labels.figure` | string | `Figure` | Caption label for `:::figure` blocks, e.g. `Abbildung`. |
| `labels.table` | string | `Table` | Caption label for `:::table` blocks, e.g. `Tabelle`. |
| `references` | list | `[]` | Bibliography entries cited with `[@key]` and listed with `:::bibliography`. |

Notes:

- `page.size` / `page.margin` only affect output if the theme references
  `{{.Page.Size}}` / `{{.Page.Margin}}` in its `@page` rule. The built-in
  `system` theme does, with fallbacks: `{{or .Page.Size "A4"}}`.
- Numbering is document-wide when enabled. `:::frontmatter` headings are
  unnumbered and omitted from the TOC by default; `:::appendix` top-level
  headings become `A`, `B`, …
- There is **no** `paginate` field — pagination is always on (paged.js). A
  `paginate:` line is silently ignored.
- Unknown frontmatter keys are ignored. Do not invent fields unless a theme reads
  them from `data`.

## References

Each entry needs either `key` or `id`; both are accepted as the citation key.
Use `text` for a raw preformatted bibliography line, otherwise structured fields
are assembled in a minimal numeric style:

```yaml
references:
  - key: lanze1982
    author: "Lanze, Werner"
    title: "Das technische Manuskript"
    year: "1982"
    publisher: "Vulkan-Verlag"
  - id: css-page
    text: "CSS Paged Media Module Level 3. https://www.w3.org/TR/css-page-3/"
```

Supported structured fields: `key`, `id`, `author`, `title`, `year`,
`publisher`, `edition`, `isbn`, `url`, `text`.

Example:

```yaml
---
mdoc: true
theme: report
title: "Q3 Engineering Review"
author: "Platform Team"
tags: [report, internal, q3]
page:
  size: A4
  margin: 25mm 22mm 28mm 22mm
numbering:
  enabled: true
labels:
  figure: "Figure"
  table: "Table"
data:
  project: Helios
  confidential: true
references:
  - key: knuth1984
    author: "Knuth, Donald E."
    title: "Literate Programming"
    year: "1984"
---
```
