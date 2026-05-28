// Package assets exposes the embedded vendor JS/CSS (paged.js, KaTeX) and the
// preview UI files (HTML/CSS/JS) as filesystems. Everything callers need to
// serve the preview or inject into print HTML lives here.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed vendor/* vendor/katex/fonts/*
var vendorEmbed embed.FS

//go:embed ui/*
var uiEmbed embed.FS

// Vendor returns a filesystem rooted at the vendor directory (so paths like
// "paged.polyfill.min.js" and "katex/katex.min.css" work directly).
func Vendor() fs.FS {
	sub, err := fs.Sub(vendorEmbed, "vendor")
	if err != nil {
		panic(err) // embed.FS sub on a directory we know exists cannot fail
	}
	return sub
}

// UI returns a filesystem rooted at the ui directory.
func UI() fs.FS {
	sub, err := fs.Sub(uiEmbed, "ui")
	if err != nil {
		panic(err)
	}
	return sub
}

// VendorBytes reads a single file out of the embedded vendor tree. Useful when
// the print pipeline needs to inline the paged.js polyfill into a generated
// HTML file.
func VendorBytes(name string) ([]byte, error) {
	return fs.ReadFile(Vendor(), name)
}

// UIBytes reads a single file out of the embedded UI tree.
func UIBytes(name string) ([]byte, error) {
	return fs.ReadFile(UI(), name)
}
