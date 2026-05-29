# Vendored web assets

These are third-party libraries committed pre-built (minified) and embedded into
the mdoc binary via `//go:embed` (see [`../assets.go`](../assets.go)).

| Path | Library | License | Upstream |
| ---- | ------- | ------- | -------- |
| `paged.min.js`, `paged.polyfill.min.js` | paged.js 0.4.3 | MIT | https://pagedjs.org |
| `katex/katex.min.js`, `katex/katex.min.css`, `katex/auto-render.min.js` | KaTeX | MIT | https://katex.org |
| `katex/fonts/KaTeX_*.woff2` | KaTeX fonts | MIT | https://github.com/KaTeX/katex-fonts |

The exact bundled versions are recorded in each file's leading `@license` banner
comment, which is preserved so the notice travels inside the compiled binary.
Full license texts are kept in [`/licenses/web/`](../../../licenses/web/) and are
shipped in the release archives.
