# mdoc themes

A theme is an HTML file processed with Go `html/template`. It receives document
metadata plus the rendered markdown body, and it must include `{{.Body}}`
somewhere to show the document.

## Selecting a theme

A `theme:` value is read one of two ways:

- **A bare key** (e.g. `theme: thesis`) names a theme in the user themes dir,
  then a built-in:
  1. `~/.config/mdoc/themes/<key>.html`
  2. built-ins: `system` and `none`

  Keys are *not* searched for next to the document, so it is always
  unambiguous which theme a key refers to. A user file overrides a built-in of
  the same key, so `~/.config/mdoc/themes/system.html` customizes the default.

- **A path** (anything with a `/`, a leading `.`/`~`, or an absolute path)
  names a theme file directly:
  - `theme: ./themes/thesis.html` — relative to the document's directory
  - `theme: ../shared/report.html` — relative paths walk up from the document
  - `theme: ~/dev/theme.html` — `~` expands to your home directory
  - `theme: /Users/me/dev.html` — absolute path

  Paths are taken verbatim, so include the `.html` extension.

Missing or broken themes (either form) are non-fatal: mdoc falls back to the
built-in `system` theme and prints a warning.

## Template data

Both the markdown body template and the theme template see:

| Field | Meaning |
|-------|---------|
| `{{.Title}}` | `title` frontmatter, default `Untitled` |
| `{{.Author}}` | `author` frontmatter, default `Anonymous` |
| `{{.Tags}}` | `tags` frontmatter |
| `{{.Page.Size}}` | `page.size` frontmatter |
| `{{.Page.Margin}}` | `page.margin` frontmatter |
| `{{.Data.<key>}}` | arbitrary values from `data` frontmatter |
| `{{.System.Date}}` | render date like `11 June 2026` |
| `{{.System.Time}}` | render time like `15:04:05` |
| `{{.System.Version}}` | mdoc version |
| `{{.Body}}` | rendered markdown HTML; theme templates only |

Use `{{or .Page.Size "A4"}}` and `{{or .Page.Margin "25mm"}}` so a document can
override page settings while the theme keeps good defaults.

## Minimal theme

```html
<style>
    @page {
        size: {{or .Page.Size "A4"}};
        margin: {{or .Page.Margin "25mm 22mm 28mm 22mm"}};
        @bottom-center { content: counter(page); }
    }

    body {
        font-family: Georgia, serif;
        font-size: 11pt;
        line-height: 1.5;
    }

    h1 { break-before: page; }
    h1:first-child { break-before: auto; }
    p { orphans: 3; widows: 3; }
    pre, table, figure { break-inside: avoid; }
</style>

{{.Body}}
```

## Stable mdoc classes

Generated apparatus uses these classes. Prefer styling these instead of relying
on fragile element positions.

| Class | Emitted for |
|-------|-------------|
| `.mdoc-matter-front` | content after `:::frontmatter` |
| `.mdoc-matter-main` | content after `:::mainmatter` |
| `.mdoc-matter-appendix` | content after `:::appendix` |
| `.mdoc-pagebreak` | `:::page` |
| `.mdoc-page-<name>` | `:::page <name>` |
| `.mdoc-secnum` | injected heading section number |
| `.mdoc-toc` | generated TOC wrapper |
| `.mdoc-toc-entry` | TOC link, with `data-level="1"` etc. |
| `.mdoc-toc-num` | TOC number |
| `.mdoc-toc-text` | TOC title |
| `.mdoc-figure` | generated figure wrapper |
| `.mdoc-table` | generated table wrapper |
| `.mdoc-figcaption` | figure/table caption |
| `.mdoc-fig-label` | injected figure label |
| `.mdoc-tab-label` | injected table label |
| `.mdoc-lof`, `.mdoc-lot` | lists of figures/tables |
| `.mdoc-lof-entry`, `.mdoc-lot-entry` | LOF/LOT links |
| `.mdoc-lof-num`, `.mdoc-lot-num` | LOF/LOT numbers |
| `.mdoc-lof-text`, `.mdoc-lot-text` | LOF/LOT captions |
| `.mdoc-xref` | number cross-reference link |
| `.mdoc-pageref` | page-reference link |
| `.mdoc-xref-unresolved` | unresolved cross-reference |
| `.mdoc-cite` | citation link |
| `.mdoc-cite-unresolved` | unresolved citation |
| `.mdoc-bib` | bibliography wrapper |
| `.mdoc-bib-entry` | bibliography item |
| `.mdoc-bib-label` | bibliography number |
| `.mdoc-bib-text` | bibliography text |

## Page numbers in generated lists

mdoc deliberately leaves TOC/LOF/LOT page numbers to the theme. Use paged.js
`target-counter`:

```css
.mdoc-toc-entry,
.mdoc-lof-entry,
.mdoc-lot-entry {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 0.5em;
}

.mdoc-toc-entry::after {
    content: target-counter(attr(href url), page);
}

.mdoc-lof-entry::after,
.mdoc-lot-entry::after {
    content: target-counter(attr(href url), page);
}

.mdoc-pageref::after {
    content: target-counter(attr(href url), page);
}
```

## Region and page-break styling

```css
.mdoc-matter-main > h1,
.mdoc-matter-appendix > h1 {
    break-before: page;
}

.mdoc-pagebreak {
    break-after: page;
}

.mdoc-page-cover {
    page: cover;
}

@page cover {
    @bottom-center { content: none; }
}
```

Use `page: <name>` on an element to select a named `@page <name>` rule. The
`:::page <name>` directive only adds `.mdoc-page-<name>`; the theme decides
whether that class selects a named page.

## Figure and table styling

```css
figure.mdoc-figure,
figure.mdoc-table {
    margin: 1em 0;
    break-inside: avoid;
}

.mdoc-figcaption {
    font-size: 0.9em;
    color: #555;
}

.mdoc-fig-label,
.mdoc-tab-label {
    font-weight: 600;
}
```

Tables inside `:::table` are still ordinary HTML `<table>` elements wrapped in
`<figure class="mdoc-table">`.

## Notes for agents

- Never omit `{{.Body}}`; an otherwise valid theme would render a blank document.
- Do not add script tags for pagination, KaTeX, or live reload; mdoc's shell
  provides those.
- Use normal CSS and paged.js page rules. Themes are trusted and raw HTML in the
  markdown body is allowed.
- There is no built-in syntax highlighting. Code fences get `language-*`
  classes only; a theme would need to provide highlighting styles/scripts.
