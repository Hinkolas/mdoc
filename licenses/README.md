# Third-party licenses

mdoc itself is licensed under the MIT License (see the top-level [`LICENSE`](../LICENSE)).
It also redistributes third-party software — both in source form in this
repository and compiled/embedded into the released binaries. This directory
collects the required copyright notices and license texts for that software.

These files are shipped inside the release archives (see `.goreleaser.yaml`) so
that binary recipients receive the notices as well.

## Bundled web assets (`web/`)

These libraries are committed pre-built and embedded into the binary via
`//go:embed` (see [`../internal/assets/`](../internal/assets/)).

| Component | Version | License | Upstream | Text |
| --------- | ------- | ------- | -------- | ---- |
| paged.js | 0.4.3 | MIT | https://pagedjs.org | [`web/pagedjs-LICENSE`](web/pagedjs-LICENSE) |
| KaTeX | bundled | MIT | https://katex.org | [`web/katex-LICENSE`](web/katex-LICENSE) |
| KaTeX fonts (`KaTeX_*.woff2`) | bundled | MIT | https://github.com/KaTeX/katex-fonts | [`web/katex-fonts-LICENSE`](web/katex-fonts-LICENSE) |

## Go module dependencies (`go/`)

Every Go module compiled into the binary has its verbatim license (and any
`NOTICE`) saved under [`go/`](go/), mirroring the module import path. See
[`go/README.md`](go/README.md) for the full list with license types.

## Regenerating

The `go/` tree is generated. After changing dependencies, refresh it with:

```sh
scripts/gen-licenses.sh
```

The `web/` texts are maintained by hand and only change when a bundled library
is upgraded.
