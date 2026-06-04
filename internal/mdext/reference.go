package mdext

import (
	"strings"

	"github.com/hinkolas/mdoc/internal/document"
)

// formatReference assembles a reference's structured fields into a single
// display string. The raw `text` escape-hatch is handled by the renderer, not
// here, so this stays pure and easy to unit-test. The v1 format is a minimal
// numeric style; richer CSL styles are future work.
func formatReference(r document.Reference) string {
	parts := make([]string, 0, 6)
	switch {
	case r.Author != "" && r.Year != "":
		parts = append(parts, r.Author+" ("+r.Year+")")
	case r.Author != "":
		parts = append(parts, r.Author)
	case r.Year != "":
		parts = append(parts, "("+r.Year+")")
	}
	if r.Title != "" {
		parts = append(parts, r.Title)
	}
	if r.Edition != "" {
		parts = append(parts, r.Edition)
	}
	if r.Publisher != "" {
		parts = append(parts, r.Publisher)
	}
	if r.ISBN != "" {
		parts = append(parts, "ISBN "+r.ISBN)
	}
	if r.URL != "" {
		parts = append(parts, r.URL)
	}
	s := strings.Join(parts, ". ")
	if s != "" && !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}
