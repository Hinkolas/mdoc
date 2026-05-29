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

`mdoc install` downloads a known-good Chromium snapshot into `$XDG_CACHE_HOME/mdoc/chromium` (`~/Library/Caches/mdoc/chromium` on macOS). If a system Chromium is already on your `PATH`, mdoc will use it as a fallback and you can skip this step.

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

Downloads Chromium into the user cache directory.

```
--chromium <revision>   pin a specific Chromium revision (default: latest known)
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

## How it works

```
foo.md  ──▶  Goldmark (GFM + footnotes)  ──▶  HTML body
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
