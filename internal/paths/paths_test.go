package paths

import (
	"path/filepath"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := map[string]RefKind{
		// flat global keys: bare words, no extension
		"thesis":          KindFlatKey,
		"system":          KindFlatKey,
		"disclaimer":      KindFlatKey,
		"contract-prefix": KindFlatKey,
		// scoped keys: contain "::"
		"kilohertz::legal::contract": KindScopedKey,
		"legal::disclaimer":          KindScopedKey,
		// paths: slash, leading dot/tilde, absolute, or a file extension
		"themes/thesis.html": KindPath,
		"./parts/intro.md":   KindPath,
		"../shared.md":       KindPath,
		"~/dev.html":         KindPath,
		"/abs/dev.html":      KindPath,
		"chapter1.md":        KindPath, // bare name WITH extension -> sibling file
		"disclaimer.md":      KindPath,
	}
	for in, want := range cases {
		if got := Classify(in); got != want {
			t.Errorf("Classify(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestScopedKeyToRelpath(t *testing.T) {
	sep := string(filepath.Separator)
	ok := map[string]string{
		"a::b::c":                    "a" + sep + "b" + sep + "c",
		"kilohertz::legal::contract": "kilohertz" + sep + "legal" + sep + "contract",
		"single":                     "single",
	}
	for in, want := range ok {
		got, err := ScopedKeyToRelpath(in)
		if err != nil {
			t.Errorf("ScopedKeyToRelpath(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ScopedKeyToRelpath(%q) = %q, want %q", in, got, want)
		}
	}

	bad := []string{
		"::leading",
		"trailing::",
		"a::::b",   // doubled separator -> empty segment
		"a::..::b", // path escape
		"a::b/c",   // slash inside a segment
	}
	for _, in := range bad {
		if _, err := ScopedKeyToRelpath(in); err == nil {
			t.Errorf("ScopedKeyToRelpath(%q) = nil error, want error", in)
		}
	}
}
