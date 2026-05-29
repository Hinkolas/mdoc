# Vendored web assets

These are third-party libraries committed pre-built (minified) and embedded into
the mdoc binary via `//go:embed` (see [`../assets.go`](../assets.go)).

| Path | Library | License | Upstream |
| ---- | ------- | ------- | -------- |
| `paged.min.js`, `paged.polyfill.min.js` | paged.js 0.4.3 | MIT | https://pagedjs.org |
| `katex/katex.min.js`, `katex/katex.min.css`, `katex/auto-render.min.js` | KaTeX | MIT | https://katex.org |
| `katex/fonts/KaTeX_*.woff2` | KaTeX fonts | MIT | https://github.com/KaTeX/katex-fonts |

Each file's leading `@license` banner records the exact bundled version and is
preserved so the notice travels inside the compiled binary.
