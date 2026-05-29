// Package theme resolves and loads HTML theme templates. Lookup order is
// project-local ./themes/<name>.html first, then the user config directory,
// then a small set of themes compiled into the binary.
//
// Resolution never hard-fails: an empty name yields the default theme, and a
// name that can't be found or won't parse falls back to the default theme
// paired with a non-fatal diagnostic error so callers can warn the user while
// still rendering something presentable.
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

// Resolve finds a theme by name. It ALWAYS returns a usable, non-nil theme.
// A project-local or user theme file takes precedence over a built-in of the
// same name, so the built-in keywords ("system", "none") double as
// overridable starting points — including the default: an empty name is
// treated as the default theme name, so dropping a themes/system.html
// customizes what unstyled-by-frontmatter documents get. A name that can't be
// found anywhere, or a theme file that fails to parse, falls back to the
// built-in default.
//
// The returned error is a non-fatal diagnostic, not a failure: callers should
// render with the returned theme and surface the error as a warning rather
// than abort. It is nil when the requested theme loaded cleanly (or resolved
// to a built-in, including when no theme was named at all).
func Resolve(name, searchProjectDir string) (*Theme, error) {
	if name == "" {
		name = DefaultName
	}

	// Theme files on disk win, so a built-in can be customized by dropping a
	// same-named file next to the document or in the user config dir.
	for _, dir := range SearchDirs(searchProjectDir) {
		candidate := filepath.Join(dir, name+".html")
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		tmpl, err := template.ParseFiles(candidate)
		if err != nil {
			return Default(), fmt.Errorf("theme %q failed to parse (%s): %w; using the built-in %q theme", name, candidate, err, DefaultName)
		}
		return &Theme{Name: name, Path: candidate, Template: tmpl}, nil
	}

	// Fall back to a built-in of that name before giving up.
	if thm := builtin(name); thm != nil {
		return thm, nil
	}

	return Default(), fmt.Errorf("theme %q not found in %s; using the built-in %q theme", name, strings.Join(SearchDirs(searchProjectDir), ", "), DefaultName)
}

// SearchDirs returns the directories searched for theme files, in order:
// the project-local <projectDir>/themes first, then the user themes dir
// (~/.config/mdoc/themes; see internal/paths).
func SearchDirs(projectDir string) []string {
	dirs := []string{filepath.Join(projectDir, "themes")}
	if userThemes, err := paths.ThemesDir(); err == nil {
		dirs = append(dirs, userThemes)
	}
	return dirs
}
