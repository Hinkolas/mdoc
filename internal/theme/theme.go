// Package theme resolves and loads HTML theme templates. A theme value is read
// one of two ways:
//
//   - A bare key (e.g. `thesis`) names a theme in the user config directory,
//     ~/.config/mdoc/themes/<key>.html, or one of the themes compiled into the
//     binary ("system", "none"). Bare keys are NOT searched for next to the
//     document — that lookup is reserved for explicit paths, so it is always
//     unambiguous which theme a key refers to.
//   - A path (anything with a "/" separator, a leading "." or "~", or an
//     absolute path) names a theme file directly. A relative path resolves from
//     the document's directory; an absolute or ~-prefixed path from the
//     filesystem root or the user's home.
//
// Resolution never hard-fails: an empty value yields the default theme, and a
// key or path that can't be found or won't parse falls back to the default
// theme paired with a non-fatal diagnostic error so callers can warn the user
// while still rendering something presentable.
package theme

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/hinkolas/mdoc/internal/paths"
)

// Theme is a parsed theme template ready to be executed by internal/render.
type Theme struct {
	Name     string
	Path     string
	Template *template.Template
}

// Fallback is the non-fatal diagnostic Resolve returns when it had to use the
// built-in default theme instead of the requested one. It implements error, so
// callers that only print err.Error() keep working (they get the full Detail);
// callers that want a terse, structured summary can type-assert and read the
// fields or call Short.
type Fallback struct {
	Requested string // theme name the document asked for
	Used      string // built-in theme used instead (DefaultName)
	Reason    string // terse reason, e.g. "not found" or "failed to parse"
	Detail    string // full human message, including searched locations
}

func (f *Fallback) Error() string { return f.Detail }

// Short returns a one-line summary suitable for a status banner, e.g.
// `"test" not found, fell back to "system"`.
func (f *Fallback) Short() string {
	return fmt.Sprintf("%q %s, fell back to %q", f.Requested, f.Reason, f.Used)
}

//go:embed system.html
var systemHTML string

const (
	// DefaultName is the theme used when a document doesn't name one: a
	// styled, dependable allrounder built into the binary. "system" reads as
	// a special keyword rather than a real on-disk theme.
	DefaultName = "system"
	// NoneName is the bare passthrough theme — the rendered body with no
	// styling at all. Opt in with `theme: none` when you want zero opinions.
	NoneName = "none"
)

// Default returns the built-in default theme. Always available, never fails.
func Default() *Theme { return builtin(DefaultName) }

// builtin returns the theme compiled into the binary for the given name, or
// nil if there's no built-in by that name. Parsing constant/embedded sources
// can't fail, so this never errors.
func builtin(name string) *Theme {
	switch name {
	case DefaultName:
		return mustParse(DefaultName, systemHTML)
	case NoneName:
		return mustParse(NoneName, "{{.Body}}")
	default:
		return nil
	}
}

func mustParse(name, src string) *Theme {
	return &Theme{Name: name, Template: template.Must(template.New(name).Parse(src))}
}

// Resolve finds a theme. It ALWAYS returns a usable, non-nil theme. The value
// is either a bare key, looked up in the user themes dir (~/.config/mdoc/themes)
// and then the built-ins, or a path, resolved relative to docDir (see
// isPath/resolvePath). An empty value is treated as the default theme key. A
// user theme file takes precedence over a built-in of the same key, so the
// built-in keywords ("system", "none") double as overridable starting points —
// including the default: dropping a ~/.config/mdoc/themes/system.html customizes
// what unstyled-by-frontmatter documents get. Anything that can't be found, or a
// theme file that fails to parse, falls back to the built-in default.
//
// The returned error is a non-fatal diagnostic, not a failure: callers should
// render with the returned theme and surface the error as a warning rather
// than abort. It is nil when the requested theme loaded cleanly (or resolved
// to a built-in, including when no theme was named at all).
func Resolve(value, docDir string) (*Theme, error) {
	if value == "" {
		value = DefaultName
	}
	if isPath(value) {
		return resolvePath(value, docDir)
	}
	return resolveKey(value)
}

// isPath reports whether a theme value is an explicit file path rather than a
// bare key. Anything with a "/" separator, a leading "." (./ or ../) or "~", or
// an absolute path is a path; a bare word like "thesis" is a key.
func isPath(value string) bool {
	return strings.ContainsRune(value, '/') ||
		strings.HasPrefix(value, ".") ||
		strings.HasPrefix(value, "~") ||
		filepath.IsAbs(value)
}

// resolveKey loads a bare-key theme from the user themes dir, then the
// built-ins. A user file overrides a same-keyed built-in.
func resolveKey(name string) (*Theme, error) {
	if userThemes, err := paths.ThemesDir(); err == nil {
		candidate := filepath.Join(userThemes, name+".html")
		if _, err := os.Stat(candidate); err == nil {
			tmpl, perr := template.ParseFiles(candidate)
			if perr != nil {
				return Default(), &Fallback{
					Requested: name,
					Used:      DefaultName,
					Reason:    "failed to parse",
					Detail:    fmt.Sprintf("theme %q failed to parse (%s): %v; using the built-in %q theme", name, paths.Display(candidate), perr, DefaultName),
				}
			}
			return &Theme{Name: name, Path: candidate, Template: tmpl}, nil
		}
	}

	if thm := builtin(name); thm != nil {
		return thm, nil
	}

	loc := "the user themes dir"
	if userThemes, err := paths.ThemesDir(); err == nil {
		loc = paths.Display(userThemes)
	}
	return Default(), &Fallback{
		Requested: name,
		Used:      DefaultName,
		Reason:    "not found",
		Detail:    fmt.Sprintf("theme %q not found in %s (use a path like ./themes/%s.html for a theme next to the document); using the built-in %q theme", name, loc, name, DefaultName),
	}
}

// resolvePath loads a theme from an explicit path. A relative path resolves
// from docDir; a "~" prefix expands to the user's home; an absolute path is used
// as-is. Unlike keys, paths are taken verbatim, so the value must include the
// .html extension.
func resolvePath(value, docDir string) (*Theme, error) {
	path := value
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[1:])
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(docDir, path)
	}
	path = filepath.Clean(path)

	if _, err := os.Stat(path); err != nil {
		return Default(), &Fallback{
			Requested: value,
			Used:      DefaultName,
			Reason:    "not found",
			Detail:    fmt.Sprintf("theme file %q not found (%s); using the built-in %q theme", value, paths.Display(path), DefaultName),
		}
	}
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return Default(), &Fallback{
			Requested: value,
			Used:      DefaultName,
			Reason:    "failed to parse",
			Detail:    fmt.Sprintf("theme file %q failed to parse (%s): %v; using the built-in %q theme", value, paths.Display(path), err, DefaultName),
		}
	}
	return &Theme{Name: value, Path: path, Template: tmpl}, nil
}

// SearchDirs returns the directories watched for bare-key theme files: the user
// themes dir (~/.config/mdoc/themes; see internal/paths). The live-preview
// watcher uses it so a global theme created or changed mid-session is noticed;
// path-valued themes are watched separately via their resolved file path.
func SearchDirs() []string {
	if userThemes, err := paths.ThemesDir(); err == nil {
		return []string{userThemes}
	}
	return nil
}
