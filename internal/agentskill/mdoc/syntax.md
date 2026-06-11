# mdoc body syntax

Three layers process the body, in order:

1. **Go `text/template`** over the raw markdown (interpolate metadata).
2. **goldmark** Markdown → HTML (GFM + footnotes; raw HTML allowed).
3. **KaTeX** renders math client-side, after pagination.

## Markdown (goldmark: GFM + footnotes)

Standard CommonMark plus GitHub extensions:

- Headings `#`…`######` (auto-assigned IDs, so anchor links work).
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

Relative URLs resolve from the **document's own directory**. Keep assets next to
the `.md` file and reference them relatively:

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
