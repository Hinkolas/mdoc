# What mdoc can't do natively yet — thesis-grade documents

Scope note. This came out of building the `thesis` theme to match a German
Bachelor-/Masterarbeit (KOMA-Script): title page, **Inhaltsverzeichnis**,
**Symbol-/Abkürzungs-/Abbildungs-/Tabellenverzeichnis**, numbered chapters,
running headers, numbered figures/tables/equations, cross-references,
bibliography, lettered appendices.

Almost all of it is *reachable* today — but only by hand-writing HTML in the
markdown body and leaning on CSS tricks. That's the wrong target. The point of
mdoc is to **write the document in markdown** and let mdoc produce the
apparatus. This file lists, feature by feature, what mdoc cannot do *natively*
today and what a bespoke mechanism could look like.

"Native" here means: the author writes markdown (plus, where unavoidable, a bit
of custom syntax, frontmatter, or a `{{…}}` directive), and mdoc generates the
rest. "Hand-HTML" means the anti-pattern below.

---

## Update — what the `mdext` extension now delivers

A goldmark extension (`internal/mdext`) has since shipped, built exactly along
the lines this document argued for (a pipeline AST pass + node renderers, not a
theme hack). It closes the worst offenders:

- **Table of contents** (gap 2) — `:::toc` collects the markdown headings and
  renders them; no hand-written `<nav>`. Exhibit A is gone.
- **Section numbering** (gap 2) — opt-in via `numbering: {enabled: true}`,
  emitted as `<span class="mdoc-secnum">` and reused by the TOC; `{.unnumbered}`,
  `{.notoc}`, `{.appendix}` (lettered) markers included.
- **Citations + bibliography** (gap 5) — `[@key]` + a frontmatter `references:`
  list + `:::bibliography`; auto-numbered by first use, with a raw `text:`
  escape-hatch. (CSL styles / `.bib` import remain future work.)
- **Figures + tables + their lists** (gap 3) — `:::figure` / `:::table` container
  directives whose markdown body carries the media and a **rich caption** (bold,
  links, `[@cite]`, `[#xref]` all work in captions); chapter-scoped auto-numbers
  and an injected `Abbildung 2.1` / `Tabelle 2.1` label; `:::lof` / `:::lot`
  render the lists. This replaced the hand-written `<figure>` / `.tablefig` /
  `<nav class="lof">` markup and the CSS figure/table counters. (Native
  sub-figure syntax is still future work — sub-figures use raw `<div>`s in the
  `:::figure` body.)
- **Cross-references** (gap 4) — `[#id]` prints a heading/figure/table number and
  links to it; `[#id page]` prints its page number (theme `target-counter`).
  This replaced the hand-typed "Abschnitt 2.2.1" and `<a class="pageref">` spans.
- **Stable heading IDs** (gap 1) — a transliterating slugifier (`ä→ae`, `ß→ss`)
  plus `{#id}` attributes; anchors are no longer lossy.
- **Document regions + page breaks** — `:::frontmatter` / `:::mainmatter` /
  `:::appendix` markers set numbering and TOC defaults per region and emit
  `<div class="mdoc-matter-…">` wrappers the theme breaks on; `:::page` forces a
  break. This replaced the hand-written `.matter-roman` / `.mainmatter` /
  `.appendix` / `<section>` layout divs — the body is now pure structure.

See the README "Generated content" section for the syntax and the `mdoc-*` CSS
class contract. `thesis.md` now uses all of these — the only hand-written HTML
left in the body is the symbol/abbreviation `<dl>` lists and the sub-figure
layout inside one `:::figure`.

**Still open** (this document's other sections still apply): running headers
(gap 7), roman→arabic page reset (gap 8), equation numbering (gap 9), PDF outline
(gap 10), native sub-figure syntax (part of gap 3), and the symbol/abbreviation
lists (nomenclature).

---

## Update — multi-file documents (`:::include`)

A thesis this size wants to live in more than one file. `:::include <path>`
splices a chapter file's body into the root before parsing (the LaTeX `\input`
model), so all the apparatus above keeps working across files from one combined
source: continuous numbering, a document-wide TOC, cross-chapter `[#id]`
references, one bibliography. The root frontmatter owns all configuration;
included files may keep their own frontmatter (discarded on include) so each
chapter stays openable on its own with `mdoc open`. `mdoc open` watches every
included file and `mdoc bundle` packs them all into the `.mdoc`.

**Known limitation — asset paths.** The combined body renders as if it all lived
in the root document's directory, so a relative `![](…)` inside an included
chapter resolves against the *root*, not the chapter's own directory. Shared
images therefore belong under the root's tree (e.g. a top-level `assets/`);
per-chapter asset directories with chapter-relative image paths are not resolved
yet. The native fix is to rewrite relative URLs in an included file to the root
when splicing (and pull those assets into the bundle) — deferred for now.

---

## The one finding that explains most of the gaps

**A theme is CSS-only. It cannot run JavaScript, and it has no view of the
document structure.** Two hard consequences:

1. `internal/render/shell.html` injects the themed HTML with
   `staging.innerHTML = html` and then hands it to paged.js. Any `<script>` in a
   theme is therefore inert (scripts set via `innerHTML` never execute). So a
   theme can't walk the headings to build a TOC, can't number anything in a way
   other parts of the document can reuse, can't post-process the page.
2. The theme only receives `{{.Body}}` (already-rendered HTML) plus scalar
   metadata (`{{.Title}}`, `{{.Data.*}}`, …). It never receives a *model* of the
   document — no list of headings, figures, tables, citations, labels.

So every "collect and number" feature — table of contents, lists of
figures/tables, cross-references, citations/bibliography, nomenclature, running
chapter titles, roman→arabic page numbering — **must be built into the mdoc
pipeline**, not a theme. There are two places mdoc can do this that a theme
cannot:

- **Go / goldmark pass** (`internal/render`): walk the parsed AST, assign
  numbers and stable IDs, collect headings/figures/tables/citations into a
  *document model*, and expose that model to the template (and inject generated
  blocks where a directive sits).
- **paged.js stage** (`shell.html`, which mdoc owns): register paged.js
  `Handler`s to do things CSS can't — restart page numbers at a region
  boundary, set running headers, emit a PDF outline. A theme can't add handlers;
  mdoc core can.

Everything below is an instance of one or both.

### Exhibit A — the anti-pattern this should kill

To get the Inhaltsverzeichnis today, the body has to contain this (every entry,
every number typed by hand; the page numbers are the only automatic part):

```html
<nav class="toc">
<span class="toc-chapter"><a href="#einleitung"><span class="t">1&emsp;Einleitung</span></a></span>
<span class="toc-section"><a href="#aufbau-des-berichtes"><span class="t">2.1&emsp;Aufbau des Berichtes</span></a></span>
<span class="toc-subsection"><a href="#uere-form"><span class="t">2.3.1&emsp;Äußere Form</span></a></span>
…18 more lines…
</nav>
```

The target is: the author writes nothing but `## Aufbau des Berichtes` in the
body and drops a single `{{toc}}` (or a theme-level placeholder) where the
Inhaltsverzeichnis belongs.

---

## Already native today (so we don't rebuild these)

| Feature | Status |
|---|---|
| GFM body, footnotes, raw HTML | ✅ goldmark |
| KaTeX math, inline + display | ✅ (but see equation numbering) |
| Justified text + hyphenation, orphan/widow control | ✅ CSS (`text-align: justify; hyphens: auto`) |
| Booktabs-style tables, two-column "definition" lists, checklist squares | ✅ CSS, from plain markdown / a little HTML |
| Page size & margins | ✅ frontmatter `page.size` / `page.margin` |
| Section/figure/table **number display** | ✅ CSS counters render `2.3.1`, `Abbildung 2.1` — **but the numbers exist only in the print output; nothing else (TOC, refs) can read them.** |
| Cross-reference **page numbers** | ✅ paged.js `target-counter(attr(href), page)` resolves "auf Seite 11" correctly |
| TOC **page numbers** | ✅ same `target-counter` — the page column of a *hand-written* TOC fills in correctly |

The CSS in `themes/thesis.html` is worth keeping as the styling layer; what's
missing is the structure to feed it.

---

## The gaps, feature by feature

Each: **Reference feature → Today → Why it's a gap → Proposed native mechanism.**

### 1. Stable, meaningful heading IDs  *(foundation for 2, 3, 6, 7)*

- **Today.** goldmark auto-IDs are lowercased, **ASCII-only**, and de-duplicated
  with a numeric suffix. Real output from this document:
  - `Äußere Form` → `uere-form`  (the `Ä`, `ü`, `ß` are dropped)
  - `Weiterführende Literatur` → `weiterfhrende-literatur`
  - a second `Einleitung` heading → `einleitung-1`
  There is no `## Title {#my-id}` attribute syntax (goldmark's Attribute
  extension isn't enabled in `internal/render/render.go`).
- **Why a gap.** Every hand-written anchor (TOC, cross-ref, list) depends on
  guessing these lossy, collision-prone slugs. It's brittle and user-hostile.
- **Proposed.** (a) Enable goldmark heading-attribute syntax so authors can pin
  `## Äußere Form {#aeussere-form}` when they want a label; (b) replace the
  default slugifier with a transliterating one (`ä→ae`, `ß→ss`, …) so
  auto-IDs are readable and stable. This unblocks everything downstream.

### 2. Section numbering as data + the table of contents

- **Reference.** `1`, `2.1`, `2.2.1`; appendices `A`, `B.1`; some headings
  unnumbered (Kurzreferat, Literaturverzeichnis). The **same** numbers appear in
  the Inhaltsverzeichnis with dot leaders and page numbers.
- **Today.** Numbers are produced by CSS counters in the theme — so they show in
  print but are **invisible to the TOC**, which is hand-written (Exhibit A),
  including re-typing every number. No concept of front-matter / main-matter /
  appendix. No way to mark a heading unnumbered except by wrapping regions in
  `<div class="mainmatter">` / `<div class="appendix">` HTML.
- **Why a gap.** Numbering lives in the rendering layer, not the document model,
  so nothing can reuse it; and the TOC has no generator at all.
- **Proposed.**
  - A **numbering engine** in the Go pass: assign chapter/section numbers with
    configurable depth, matter regions, and appendix lettering; expose each
    heading as `{number, text, id, level, matter}` in the document model.
    Drive regions/unnumbered from markdown markers or frontmatter (e.g. a
    `:::frontmatter` / `:::appendix` fence, or `#* Unnumbered`).
  - A **`{{toc}}` directive** (or a theme placeholder like `{{.TOC}}`) that mdoc
    replaces with a generated nav using those numbers + IDs. Page numbers keep
    coming from paged.js `target-counter` at render time (already works); the
    theme just styles `.toc`. Dot leaders: ship a CSS helper, or have mdoc inject
    a paged.js leader handler (CSS `leader()` is unsupported — see Constraints).

### 3. First-class figures/tables + their lists (Abbildungs-/Tabellenverzeichnis)

- **Reference.** `Abbildung 2.1: …` under figures, `Tabelle 2.1: …` above tables,
  numbered per chapter, each collected into its own list with page numbers.
- **Today.** A captioned, numbered figure needs raw
  `<figure id="…"><img><figcaption>…</figcaption></figure>`; a captioned table
  needs a hand-rolled wrapper; sub-figures must be plain `<div>`s to stay out of
  the counter. The two lists are hand-written, same as the TOC. (The README
  roadmap already flags "first-class figure syntax" and "auto figure index.")
- **Why a gap.** No figure/table shorthand, no auto-numbering visible to a list,
  no list generator.
- **Proposed.** A figure shorthand (`![alt](src "caption"){#fig:x}` →
  `<figure>` with id + auto-numbered caption), the same for tables, and
  `{{lof}}` / `{{lot}}` (or `{{.Figures}}` / `{{.Tables}}`) directives that emit
  the lists from the collected model.

### 4. Cross-references by label (`siehe Abschnitt 2.2.1 auf Seite 8`)

- **Today.** The **page** is automatic (`<a class="pageref" href="#id">` +
  `target-counter`), but the **section number** ("2.2.1") and the anchor are
  typed by hand against a lossy slug.
- **Proposed.** A reference syntax — `[@sec:aufbau]` or `{{ref "aufbau"}}` —
  that resolves to the number (from the numbering engine, #2) and/or the page
  (from `target-counter`). Depends on #1 (labels) and #2 (numbers).

### 5. Citations + bibliography (Literaturverzeichnis, `[1]`…`[9]`)

- **Reference.** Numbered references, cited in text as `[3]`, auto-numbered by
  first citation, formatted to a style (DIN 1505 here).
- **Today.** Hand-written `<ol class="bibliography">` with `<li id="bibN">`, and
  `<a href="#bibN">[n]</a>` in the text. No citation processor, no `.bib`, no
  auto-numbering, no sorting, no styles.
- **Proposed.** A reference source (frontmatter list or a `.bib`/CSL-JSON file),
  a cite syntax `[@key]`, and a `{{bibliography}}` directive that emits a
  numbered/sorted list. Realistically wrap a Go CSL/citeproc library so styles
  (DIN, IEEE, APA) are config, not code.

### 6. Symbol & abbreviation lists (Symbol-/Abkürzungsverzeichnis / nomenclature)

- **Reference.** Alphabetically sorted lists of symbols and abbreviations used in
  the text (LaTeX `nomencl`/`glossaries`).
- **Today.** Hand-written `<dl>`. No collect-from-use, no sorting, no
  first-use expansion.
- **Proposed.** A define/collect mechanism — e.g. `{{abbr "DIN" "Deutsches
  Institut für Normung"}}` expands on first use and registers the entry — plus
  `{{abbreviations}}` / `{{symbols}}` directives that emit sorted lists. Or a
  frontmatter table for the simple case.

### 7. Running headers (chapter title in the top margin)

- **Reference.** Continuation pages show the current chapter (italic, ruled).
- **Today — does not work from a theme.** paged.js named strings fight us on
  every front (all confirmed empirically, see Constraints): `string-set` is
  silently dropped on any element that *also* carries
  `counter-increment`/`counter-reset` (so it can't ride on the numbered `h1`);
  the value is carried with a one-page lag (opening pages show the *previous*
  chapter); `counter(name, upper-alpha)`, `content(before)`, and
  `string(name, first)` are all rejected by this build. The theme currently
  ships **footer-only** for this reason.
- **Proposed.** mdoc precomputes each heading's display string ("2  Title") and
  emits a dedicated, counter-free `string-set` carrier element — or, cleaner,
  injects a paged.js running-header `Handler` in `shell.html`. (Or upgrade the
  bundled paged.js.) Then a theme just writes `string(chaptertitle)`.

### 8. Roman front matter → arabic main matter, restarting at 1

- **Reference.** Title unnumbered; abstract/declaration `i, ii, iii`; the
  Inhaltsverzeichnis restarts at `1` and the body continues.
- **Today — impossible from CSS.** A mid-document `counter-reset: page` does
  **not** stick: the reset page shows `0` and the following pages resume the
  global count (verified with a 4-page probe). The theme uses continuous decimal
  numbering as the only reliable option.
- **Proposed.** mdoc injects a paged.js page-counter `Handler` (it owns
  `shell.html`) to reset/relabel page numbers at matter boundaries, configured
  from frontmatter (e.g. `numbering: {frontmatter: roman, mainmatter: {restart:
  1}}`). A theme cannot register handlers; mdoc core can.

### 9. Numbered equations + equation references

- **Reference.** `(2.1)`, `(2.2a)` right-aligned; referenced by number.
- **Today.** Manual `\tag{2.1}` inside the math. KaTeX auto-render does not
  number equations or resolve `\label`/`\ref`.
- **Proposed.** An equation-numbering pass (inject `\tag` from a chapter-aware
  counter, keep a label→number map) and `[@eq:x]` references. Lower priority.

### 10. PDF outline / bookmarks

- **Reference PDF** carries an `/Outlines` tree (clickable bookmarks).
- **Today.** mdoc's output has **no `/Outlines`** (it does emit `/Dest` named
  destinations, so internal links work — but no bookmark sidebar). Verified by
  inspecting both PDFs.
- **Proposed.** Build an outline from the heading model and emit it in the
  Chromium print step (mdoc owns the print pipeline), or post-process the PDF.

### 11. Title / cover page  *(addressed in this theme)*

- **Today.** The `thesis` theme renders the cover itself from frontmatter
  `data` (university, faculty, subtitle, …) plus `title`/`author` — the
  theme is an `html/template` with full access to those fields, so there is no
  cover HTML in the document body at all. (The logo is the exception: a theme
  can't read a file from a path — `html/template` has no file-reading helper —
  so the mark is embedded as inline SVG in the theme rather than referenced via
  `data.logo`.) Any theme can do the same; a conventional `cover:` schema in
  core would only standardise the field names.

---

## Rendering-engine constraints we discovered (reference facts)

Hard facts about the bundled paged.js / goldmark, learned the hard way. The
"workaround" column says whether **mdoc core** (not a theme) could neutralize it.

| # | Constraint | Impact | mdoc can work around? |
|---|---|---|---|
| C1 | `target-counter(attr(href), page)` **works** | cross-ref + TOC page numbers are solvable | — (already usable) |
| C2 | CSS counters **work** | section/figure/table number *display* is solvable | — |
| C3 | CSS `leader('.')` **unsupported** | dotted TOC leaders need a CSS hack (opaque-label + absolute dotted rule) | inject a leader handler, or ship the CSS helper |
| C4 | mid-document `counter-reset: page` **doesn't stick** | no roman→arabic restart from CSS (gap 8) | yes — paged.js page-counter handler |
| C5 | `string-set` is **dropped** when the element also has `counter-increment`/`counter-reset` | running header can't ride the numbered `h1` (gap 7) | yes — counter-free carrier element or handler |
| C6 | `string-set` rejects 2-arg `counter(n, upper-alpha)` and `content(before)`; `string(name, first)` rejected | no chapter letter / no "current section on its own page" in headers | yes — precompute strings in Go |
| C7 | `counter-increment` on `::before` is **ignored** | counters must increment on the element, not the caption pseudo | — (just author CSS accordingly) |
| C8 | theme `<script>` **never executes** (`innerHTML` injection in `shell.html`) | no theme-side generation at all | this is the root reason work must live in mdoc |
| C9 | goldmark auto-IDs: lowercase, **ASCII-only**, dedupe `-1`; no `{#id}` syntax | brittle anchors (gap 1) | yes — slugifier + heading-attribute extension |

C8 is the headline: it's *why* the TOC/lists/refs/headers can't be theme
features and must be pipeline features.

---

## Suggested build order

A small foundation makes most of the list collapse into one-line directives.

1. **Foundation**
   - Stable heading IDs: transliterating slugifier + `{#id}` attribute syntax (gap 1, C9).
   - A **document model** built during the Go pass: headings (with numbers,
     matter region, level, id), figures, tables, equations, labels, citations.
   - A **numbering engine** over that model (gap 2): depth, matter regions,
     unnumbered headings, appendix lettering.
   - Expose the model to the template, and support body **directives** (`{{toc}}`,
     `{{lof}}`, `{{lot}}`, `{{bibliography}}`, `{{abbreviations}}`, `{{ref …}}`)
     that mdoc resolves before/around goldmark.
2. **Then these become easy**
   - `{{toc}}`, `{{lof}}`, `{{lot}}` (gaps 2, 3) — numbers from the engine, page
     numbers from `target-counter` (C1).
   - `[@label]` cross-references (gap 4).
   - Figure/table shorthand (gap 3).
3. **paged.js handlers in `shell.html`** (mdoc-owned, not theme)
   - Page-number restart / roman front matter (gap 8, C4).
   - Running headers (gap 7, C5/C6).
   - PDF outline from the heading model (gap 10).
4. **Bigger, optional**
   - Citations + bibliography styles via CSL (gap 5).
   - Nomenclature collect-from-use (gap 6).
   - Equation numbering + refs (gap 9).

---

## The proof-of-concept (`thesis.html` + `thesis.md`)

`themes/thesis.html` and `thesis.md` render a 19-page thesis that matches the
reference layout closely: title page, the four Verzeichnisse, numbered chapters,
booktabs table, numbered figure + sub-figures, numbered equation, cross-refs
with page numbers, bibliography, lettered appendices, checklists.

Treat it as two things: a **design/CSS reference** for the eventual native
features (the styling is reusable once mdoc generates the structure), and a
**demonstration of the hand-authoring tax** — the TOC, the lists, the
bibliography, the numbered figures and the symbol/abbreviation lists are all
hand-written HTML today. That tax is exactly what the mechanisms above remove.
