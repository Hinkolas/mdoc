# mdoc CLI

Most commands take a root markdown file. The root may pull chapter files in via
`:::include`; `mdoc open` watches included files and `mdoc bundle` stores them.

## `mdoc print <file>` — render to PDF

```bash
mdoc print report.md              # writes report.pdf next to the source
mdoc print report.md -o out.pdf   # custom output path
mdoc print report.md --html       # also write the rendered .html alongside
mdoc print report.md --force      # overwrite an existing output file
```

- `-o, --output <path>` — output PDF path (default `<input>.pdf`).
- `--html` — also write the intermediate rendered HTML.
- `-f, --force` — overwrite an existing output file without prompting.
- In a TTY it prints a summary banner; in a pipe it prints only the output path
  (so `mdoc print x.md | xargs open` works).

## `mdoc open <file>` — live preview

```bash
mdoc open report.md          # chromeless Chromium window, hot-reloads on save
mdoc open report.md -p 0     # pick a free port instead of the default 7768
mdoc open report.md --verbose
```

- `-p, --port <n>` — preview server port (default `7768`, `0` = free port).
- `--verbose` — stream reload and theme diagnostic logs to the terminal.
- Watches the root document, included files, theme search directories, and the
  active theme; edits re-render with no flicker.

## `mdoc bundle <file>` — portable bundle

```bash
mdoc bundle report.md            # writes report.mdoc (a zip) next to the source
mdoc bundle report.md -o b.mdoc
mdoc bundle report.md --force
```

- Bundles the document, its theme, referenced assets, and any `:::include`d
  chapter files (at their relative paths) into a `.mdoc` zip.
- `-o, --output <path>` — default `<input>.mdoc`.
- `-f, --force` — overwrite an existing bundle without prompting.
- Included files must live under the root document directory to bundle cleanly.

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

## `mdoc uninstall` — remove mdoc

```bash
mdoc uninstall            # remove binary, skill, and cache; keep config
mdoc uninstall --purge    # also remove ~/.config/mdoc; skip all prompts
```

Removes the mdoc binary, the bundled skill (`~/.claude/skills/mdoc` and
`~/.codex/skills/mdoc`), and the Chromium cache. The config directory
(`~/.config/mdoc`, themes and custom CSS) is kept by default. On an interactive
terminal it asks whether to remove the config directory, then confirms before
deleting; `--purge` removes everything without prompting. Skills installed to a
custom `--path` aren't tracked and must be removed by hand.

## Themes (important for authoring)

See `themes.md` for key-vs-path selection, template data, generated classes,
page numbers, and starter CSS. In short: omit `theme` for the built-in `system`
theme, use `theme: none` for bare HTML, name a global theme by key
(`theme: report` → `~/.config/mdoc/themes/report.html`), or point at a file next
to the document by path (`theme: ./themes/report.html`) — copy
`examples/plain.html` as a starting point.
