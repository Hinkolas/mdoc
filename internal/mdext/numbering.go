package mdext

import (
	"strconv"
	"strings"

	"github.com/hinkolas/mdoc/internal/document"
)

// numbering.go is the section-number format engine. A heading's number is built
// from the live per-level counters via a small template language so themes and
// documents can shape it freely — custom text around the number ("§5"), a
// numbering system per level (decimal, roman, letters), or a mix ("5.a") — while
// the number stays a plain Go string the rest of the pipeline (TOC, `[#id]`
// cross-references, the bibliography) can read. With no per-level config it
// reproduces the historical default: decimal, dot-joined, appendix lettering.

// renderNumber formats a heading's section number from the live counters using
// the per-level template and style in cfg, falling back to the built-in default
// for any unconfigured level. The result is the bare number content ("§5",
// "1.1", "A.1") — the separating space between the number and the heading title
// is added by the SecNum renderer, so trailing whitespace in a template is
// redundant and trimmed. Keeping the number bare also keeps it clean where it is
// reused without a following title: TOC entries and `[#id]` cross-references.
func renderNumber(cfg document.Numbering, counters []int, level int, appendix bool) string {
	out := expandTemplate(levelTemplate(cfg, level), func(n int) string {
		if n < 1 || n >= len(counters) {
			return ""
		}
		return renderCounter(counters[n], levelStyle(cfg, n, appendix))
	})
	return strings.TrimRight(out, " \t")
}

// levelNumbered resolves whether a heading level is numbered: an explicit
// per-level `enabled` wins, otherwise the global Numbering.Enabled applies. A
// level set to false stays unnumbered even when numbering is otherwise on.
func levelNumbered(cfg document.Numbering, level int) bool {
	if lv, ok := cfg.Levels[levelKey(level)]; ok && lv.Enabled != nil {
		return *lv.Enabled
	}
	return cfg.Enabled
}

// levelKey is the Levels map key for a heading level: 1 -> "h1".
func levelKey(level int) string { return "h" + strconv.Itoa(level) }

// levelTemplate returns the configured template for a level, or the default
// "{1}.{2}…{level}" (dot-joined counters up to the level) when none is set.
func levelTemplate(cfg document.Numbering, level int) string {
	if lv, ok := cfg.Levels[levelKey(level)]; ok && lv.Template != "" {
		return lv.Template
	}
	parts := make([]string, 0, level)
	for i := 1; i <= level; i++ {
		parts = append(parts, "{"+strconv.Itoa(i)+"}")
	}
	return strings.Join(parts, ".")
}

// levelStyle returns the numbering style for a level's counter. The appendix
// region forces upper-alpha lettering on the chapter (level 1), matching the
// historical default, regardless of any configured style; deeper levels and the
// non-appendix case use the configured style, defaulting to decimal.
func levelStyle(cfg document.Numbering, level int, appendix bool) string {
	if appendix && level == 1 {
		return "upper-alpha"
	}
	if lv, ok := cfg.Levels[levelKey(level)]; ok && lv.Style != "" {
		return lv.Style
	}
	return "decimal"
}

// expandTemplate replaces every `{n}` placeholder (n a single digit 1..6) in
// tmpl with counter(n) and passes all other text through verbatim. An
// unrecognised brace group (e.g. `{0}`, `{x}`, `{12}`) is left as literal text.
func expandTemplate(tmpl string, counter func(n int) string) string {
	var b strings.Builder
	for i := 0; i < len(tmpl); {
		if tmpl[i] == '{' {
			if j := strings.IndexByte(tmpl[i:], '}'); j > 1 {
				if n, err := strconv.Atoi(tmpl[i+1 : i+j]); err == nil && n >= 1 && n <= 6 {
					b.WriteString(counter(n))
					i += j + 1
					continue
				}
			}
		}
		b.WriteByte(tmpl[i])
		i++
	}
	return b.String()
}

// renderCounter renders a single counter value in the given CSS-style numbering
// system. Unknown styles fall back to decimal.
func renderCounter(n int, style string) string {
	switch style {
	case "lower-alpha":
		return alpha(n, 'a')
	case "upper-alpha":
		return alpha(n, 'A')
	case "lower-roman":
		return strings.ToLower(roman(n))
	case "upper-roman":
		return roman(n)
	default: // "decimal" and anything unrecognised
		return strconv.Itoa(n)
	}
}

// alpha renders n (1-based) as a bijective base-26 letter sequence starting at
// base: 1->A, 26->Z, 27->AA, … Non-positive values fall back to decimal.
func alpha(n int, base rune) string {
	if n < 1 {
		return strconv.Itoa(n)
	}
	var out []rune
	for n > 0 {
		n--
		out = append([]rune{base + rune(n%26)}, out...)
		n /= 26
	}
	return string(out)
}

// roman renders n as an upper-case Roman numeral. Values outside 1..3999 (which
// Roman numerals can't represent) fall back to decimal.
func roman(n int) string {
	if n < 1 || n > 3999 {
		return strconv.Itoa(n)
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	for i, v := range vals {
		for n >= v {
			b.WriteString(syms[i])
			n -= v
		}
	}
	return b.String()
}
