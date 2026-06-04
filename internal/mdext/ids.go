package mdext

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// transIDs implements parser.IDs. It behaves like goldmark's default id
// generator (lowercase, keep [a-z0-9], spaces/-/_ → '-', dedupe with a numeric
// suffix) but transliterates common non-ASCII letters first, so "Äußere Form"
// becomes "aeussere-form" instead of the default's lossy "uere-form".
type transIDs struct {
	values map[string]bool
}

// NewIDs returns a transliterating parser.IDs. Pass it via
// parser.NewContext(parser.WithIDs(NewIDs())) at Convert time.
func NewIDs() parser.IDs {
	return &transIDs{values: map[string]bool{}}
}

func (s *transIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	folded := fold(value)
	result := make([]byte, 0, len(folded))
	for _, v := range folded {
		switch {
		case v >= 'A' && v <= 'Z':
			result = append(result, v+('a'-'A'))
		case (v >= 'a' && v <= 'z') || (v >= '0' && v <= '9'):
			result = append(result, v)
		case v == ' ' || v == '\t' || v == '\n' || v == '-' || v == '_':
			result = append(result, '-')
		}
	}
	if len(result) == 0 {
		if kind == ast.KindHeading {
			result = []byte("heading")
		} else {
			result = []byte("id")
		}
	}
	if !s.values[string(result)] {
		s.values[string(result)] = true
		return result
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", result, i)
		if !s.values[candidate] {
			s.values[candidate] = true
			return []byte(candidate)
		}
	}
}

func (s *transIDs) Put(value []byte) {
	s.values[string(value)] = true
}

// fold transliterates a UTF-8 string to ASCII, mapping common Latin diacritics
// and German ligatures to their base letters and dropping anything else
// non-ASCII. ASCII bytes pass through unchanged.
func fold(value []byte) []byte {
	out := make([]byte, 0, len(value))
	for _, r := range string(value) {
		if r < 0x80 {
			out = append(out, byte(r))
			continue
		}
		if rep, ok := foldMap[r]; ok {
			out = append(out, rep...)
		}
		// unknown non-ASCII runes are dropped
	}
	return out
}

var foldMap = map[rune]string{
	'ä': "ae", 'ö': "oe", 'ü': "ue", 'ß': "ss",
	'Ä': "Ae", 'Ö': "Oe", 'Ü': "Ue",
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'å': "a",
	'À': "A", 'Á': "A", 'Â': "A", 'Ã': "A", 'Å': "A",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'È': "E", 'É': "E", 'Ê': "E", 'Ë': "E",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'Ì': "I", 'Í': "I", 'Î': "I", 'Ï': "I",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ø': "o",
	'Ò': "O", 'Ó': "O", 'Ô': "O", 'Õ': "O", 'Ø': "O",
	'ù': "u", 'ú': "u", 'û': "u",
	'Ù': "U", 'Ú': "U", 'Û': "U",
	'ñ': "n", 'Ñ': "N", 'ç': "c", 'Ç': "C", 'ý': "y", 'ÿ': "y",
}
