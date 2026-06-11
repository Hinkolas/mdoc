# mdoc CLI

Each command takes a single markdown file.

## `mdoc print <file>` — render to PDF

```bash
mdoc print report.md              # writes report.pdf next to the source
mdoc print report.md -o out.pdf   # custom output path
mdoc print report.md --html       # also write the rendered .html alongside
```

- `-o, --output <path>` — output PDF path (default `<input>.pdf`).
- `--html` — also write the intermediate rendered HTML.
- In a TTY it prints a summary banner; in a pipe it prints only the output path
  (so `mdoc print x.md | xargs open` works).

## `mdoc open <file>` — live preview

```bash
mdoc open report.md          # chromeless Chromium window, hot-reloads on save
mdoc open report.md -p 0     # pick a free port instead of the default 7768
```

- `-p, --port <n>` — preview server port (default `7768`, `0` = free port).
- Watches both the document and its theme; edits re-render with no flicker.

## `mdoc bundle <file>` — portable bundle

```bash
mdoc bundle report.md            # writes report.mdoc (a zip) next to the source
mdoc bundle report.md -o b.mdoc
```

- Bundles the document, its theme, referenced assets, and any `:::include`d
  chapter files (at their relative paths) into a `.mdoc` zip.
- `-o, --output <path>` — default `<input>.mdoc`.

## `mdoc install` — setup wizard

```bash
mdoc install                         # interactive setup wizard
mdoc install --chromium              # download a known-good Chromium snapshot
mdoc install --chromium=<rev>        # pin a specific revision
mdoc install --skill claude          # install this skill for Claude
mdoc install --skill codex           # install this skill for Codex
mdoc install --skill all             # install this skill for Claude and Codex
mdoc install --skill claude --path ~/agent/skills
```

Rendering needs Chromium. With no flags in an interactive terminal,
`mdoc install` asks whether to use an existing packaged/system Chromium or
download mdoc's packaged snapshot. It then asks whether to install this bundled
skill for Claude, Codex, both, or neither.

In scripts/non-interactive terminals, plain `mdoc install` preserves the old
behavior and downloads a known-good Chromium snapshot into the user cache dir.
If a system Chromium is already on `PATH`, mdoc can use it as a fallback.

The `--skill` flag skips the wizard and copies the bundled skill directly.
Claude installs to `~/.claude/skills/mdoc`; Codex installs to
`~/.codex/skills/mdoc`. `--path <dir>` overrides the parent skills directory
for a single target, so `--path ~/agent/skills` writes
`~/agent/skills/mdoc`.

## Themes (important for authoring)

`theme: <name>` resolves to the first match of:

1. `<document-dir>/themes/<name>.html`
2. `~/.config/mdoc/themes/<name>.html`
3. a **built-in keyword**: `system` (a styled, dependable allrounder — the
   default when `theme` is omitted/empty) or `none` (bare rendered body, no
   styling). A same-named theme file on disk overrides either built-in,
   including the default — drop a `themes/system.html` to restyle everything.

A missing or broken theme is **not fatal**: if a named `theme:` can't be found
(or the file won't parse), mdoc falls back to `system` and prints a warning —
`mdoc print`/`bundle` warn on stderr and still produce output; `mdoc open`
warns on the terminal and shows the warning in the preview UI's status pill.
When authoring:

- Omit `theme` (or `theme: system`) for the styled default; `theme: none` for
  a bare body, or
- Copy `examples/plain.html` (in this skill) to `themes/<name>.html` next to
  the document and customize its CSS.

A theme is a Go HTML template that receives the document metadata and the
rendered body as `{{.Body}}`. It typically wraps a `<style>` block — with
paged.js `@page` rules for size/margins/page numbers — around `{{.Body}}`. See
`examples/plain.html` for a complete, working example to copy.
