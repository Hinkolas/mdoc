---
mdoc: true
theme: system
title: "A Book in Chapters"
author: "Nicholas Hinke"
tags: [example, include, multi-file]
numbering:
  enabled: true
---

:::frontmatter

# Preface {.unnumbered}

This short book is split across several files — one per chapter — and stitched
back together here with `:::include`. The root document owns all configuration
(theme, numbering, page) and decides the overall flow; each chapter file is a
plain markdown body that can also be opened on its own with
`mdoc open chapters/01-introduction.md`.

Because includes are spliced before the document is parsed, everything that
spans chapters just works: the heading numbering below runs continuously, the
table of contents covers every chapter, and a cross-reference can point from one
chapter into another.

:::toc

:::mainmatter

:::include chapters/01-introduction.md

:::include chapters/02-methods.md
