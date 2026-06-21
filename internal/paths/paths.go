// Package paths resolves mdoc's per-user directories.
//
// User-authored files (themes, and any future hand-edited config) live under
// an XDG-style config directory — $XDG_CONFIG_HOME/mdoc when set, otherwise
// ~/.config/mdoc — on every platform. That's deliberately the same friendly
// ~/.config path everywhere rather than os.UserConfigDir's platform default
// (e.g. ~/Library/Application Support on macOS), since users have to drop
// theme files in by hand and navigating to Library isn't ergonomic.
//
// Regeneratable downloads (the Chromium snapshot) are NOT here — they stay in
// the system cache dir; see internal/browser.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigDir returns mdoc's user config directory: $XDG_CONFIG_HOME/mdoc when
// that env var is set, otherwise ~/.config/mdoc.
func ConfigDir() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "mdoc"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "mdoc"), nil
}

// ThemesDir returns the user-level themes directory, <ConfigDir>/themes.
func ThemesDir() (string, error) {
	cfg, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "themes"), nil
}

// IncludesDir returns the user-level includes directory, <ConfigDir>/includes.
// It holds reusable markdown partials referenced by bare or scoped `:::include`
// keys (the include analogue of ThemesDir).
func IncludesDir() (string, error) {
	cfg, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "includes"), nil
}

// RefKind classifies how a theme/include reference string should be resolved.
// The three forms are mutually exclusive and decided by Classify.
type RefKind int

const (
	// KindFlatKey is a bare word (e.g. "thesis"): a global asset looked up by
	// name in the relevant config dir (themes/ or includes/).
	KindFlatKey RefKind = iota
	// KindScopedKey contains "::" (e.g. "kilohertz::legal::contract"): a global
	// asset in a subdirectory of the config dir.
	KindScopedKey
	// KindPath is an explicit filesystem path (contains "/", a leading "." or
	// "~", is absolute, or carries a file extension): resolved relative to the
	// document, not the config dir.
	KindPath
)

// Classify decides how a reference value is resolved. The "::" scope check comes
// first so a scoped key is never mistaken for anything else; otherwise any path
// sigil — a "/" separator, a leading "." or "~", an absolute path, or a file
// extension — marks a path, and a plain bare word is a flat global key. The file
// extension is what distinguishes a global key ("disclaimer") from a sibling file
// referenced by bare name ("chapter1.md"). This rule is shared by theme and
// include resolution so both behave identically.
func Classify(value string) RefKind {
	switch {
	case strings.Contains(value, "::"):
		return KindScopedKey
	case strings.ContainsRune(value, '/'),
		strings.HasPrefix(value, "."),
		strings.HasPrefix(value, "~"),
		filepath.IsAbs(value),
		filepath.Ext(value) != "":
		return KindPath
	default:
		return KindFlatKey
	}
}

// ScopedKeyToRelpath turns a "::"-scoped key into a relative filesystem path,
// e.g. "kilohertz::legal::contract" -> "kilohertz/legal/contract". It rejects
// malformed scopes — an empty segment (from a leading, trailing, or doubled
// "::") or a "." / ".." segment that would escape the config dir — so a bad key
// fails loudly instead of resolving somewhere surprising.
func ScopedKeyToRelpath(value string) (string, error) {
	segments := strings.Split(value, "::")
	for _, seg := range segments {
		if seg == "" {
			return "", fmt.Errorf("malformed scoped key %q: empty segment (check for a leading, trailing, or doubled \"::\")", value)
		}
		if seg == "." || seg == ".." || strings.ContainsAny(seg, "/\\") {
			return "", fmt.Errorf("malformed scoped key %q: invalid segment %q", value, seg)
		}
	}
	return filepath.Join(segments...), nil
}

// Display formats a path for showing to the user: made absolute, then with the
// home directory collapsed to "~" (e.g. "~/Github/mdoc/doc.md"). Paths outside
// the home directory, or anything that can't be resolved, are returned as the
// best absolute form available. This is presentation only — never feed the
// result back into file operations.
func Display(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return abs
	}
	if abs == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(abs, home+string(filepath.Separator)); ok {
		return "~" + string(filepath.Separator) + rest
	}
	return abs
}
