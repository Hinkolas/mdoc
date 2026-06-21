package mdext

import (
	"testing"

	"github.com/hinkolas/mdoc/internal/document"
)

func ptr(b bool) *bool { return &b }

// counters builds a 7-slot counter array (indices 1..6) from the given values.
func counters(vals ...int) []int {
	c := make([]int, 7)
	copy(c[1:], vals)
	return c
}

func TestRenderCounterStyles(t *testing.T) {
	cases := []struct {
		n     int
		style string
		want  string
	}{
		{5, "decimal", "5"},
		{1, "lower-alpha", "a"},
		{26, "lower-alpha", "z"},
		{27, "lower-alpha", "aa"},
		{1, "upper-alpha", "A"},
		{28, "upper-alpha", "AB"},
		{4, "lower-roman", "iv"},
		{9, "upper-roman", "IX"},
		{2024, "upper-roman", "MMXXIV"},
		{7, "", "7"},            // empty -> decimal
		{7, "bogus", "7"},       // unknown -> decimal
		{0, "upper-alpha", "0"}, // out of range -> decimal fallback
	}
	for _, c := range cases {
		if got := renderCounter(c.n, c.style); got != c.want {
			t.Errorf("renderCounter(%d, %q) = %q, want %q", c.n, c.style, got, c.want)
		}
	}
}

func TestRenderNumberDefault(t *testing.T) {
	// With no per-level config the engine must reproduce the historical default:
	// decimal, dot-joined, with appendix lettering on the chapter.
	cfg := document.Numbering{Enabled: true}
	cases := []struct {
		counters []int
		level    int
		appendix bool
		want     string
	}{
		{counters(1), 1, false, "1"},
		{counters(1, 1), 2, false, "1.1"},
		{counters(2, 3, 4), 3, false, "2.3.4"},
		{counters(1), 1, true, "A"},
		{counters(2, 1), 2, true, "B.1"},
	}
	for _, c := range cases {
		if got := renderNumber(cfg, c.counters, c.level, c.appendix); got != c.want {
			t.Errorf("renderNumber(level=%d, appendix=%v) = %q, want %q", c.level, c.appendix, got, c.want)
		}
	}
}

func TestRenderNumberCustomTemplateAndStyle(t *testing.T) {
	cfg := document.Numbering{
		Enabled: true,
		Levels: map[string]document.NumLevel{
			"h1": {Template: "§{1}", Style: "decimal"},
			"h2": {Template: "{1}.{2}", Style: "lower-alpha"},
			"h3": {Template: "{3}", Style: "lower-roman"},
		},
	}
	cases := []struct {
		counters []int
		level    int
		want     string
	}{
		{counters(5), 1, "§5"},        // custom prefix text
		{counters(5, 2), 2, "5.b"},    // mixed: decimal . lower-alpha
		{counters(5, 2, 3), 3, "iii"}, // local-only roman
	}
	for _, c := range cases {
		if got := renderNumber(cfg, c.counters, c.level, false); got != c.want {
			t.Errorf("renderNumber(level=%d) = %q, want %q", c.level, got, c.want)
		}
	}
}

func TestLevelNumbered(t *testing.T) {
	cfg := document.Numbering{
		Enabled: true,
		Levels: map[string]document.NumLevel{
			"h3": {Enabled: ptr(false)},
		},
	}
	if !levelNumbered(cfg, 1) {
		t.Error("h1 should inherit enabled=true")
	}
	if levelNumbered(cfg, 3) {
		t.Error("h3 explicitly disabled should be unnumbered")
	}

	off := document.Numbering{
		Enabled: false,
		Levels: map[string]document.NumLevel{
			"h1": {Enabled: ptr(true)},
		},
	}
	if !levelNumbered(off, 1) {
		t.Error("h1 explicitly enabled should number even when global is off")
	}
	if levelNumbered(off, 2) {
		t.Error("h2 should inherit global off")
	}
}

func TestExpandTemplateLeavesUnknownBracesLiteral(t *testing.T) {
	got := expandTemplate("{1}-{x}-{9}", func(n int) string { return "N" })
	if want := "N-{x}-{9}"; got != want {
		t.Errorf("expandTemplate = %q, want %q", got, want)
	}
}
