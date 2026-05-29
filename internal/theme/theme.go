// Package theme resolves and loads HTML theme templates. Lookup order is
// project-local ./themes/<name>.html first, then the user config directory.
//
// Resolution never hard-fails: an empty name, a name that can't be found, or
// a theme file that won't parse all fall back to the built-in minimal Default
// theme. When a named theme couldn't be loaded the fallback is paired with a
// non-fatal diagnostic error so callers can warn the user while still
// rendering something.
package theme

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// Theme is a parsed theme template ready to be executed by internal/render.
type Theme struct {
	Name     string
	Path     string
	Template *template.Template
}

// DefaultName is the name reported for the built-in fallback theme.
const DefaultName = "default"

// Default returns the built-in minimal theme: it emits the rendered body with
// no styling. It is always available and never fails to load.
func Default() *Theme {
	// Parsing a constant template cannot fail.
	tmpl := template.Must(template.New(DefaultName).Parse("{{.Body}}"))
	return &Theme{Name: DefaultName, Template: tmpl}
}

// Resolve finds a theme by name. It ALWAYS returns a usable, non-nil theme.
// An empty name, a name that can't be found in any search directory, or a
// theme file that fails to parse all fall back to the built-in Default theme.
//
// The returned error is a non-fatal diagnostic, not a failure: callers should
// render with the returned theme and surface the error as a warning rather
// than abort. It is nil when the requested theme loaded cleanly (or when no
// theme was requested at all).
func Resolve(name, searchProjectDir string) (*Theme, error) {
	if name == "" {
		return Default(), nil
	}

	for _, dir := range SearchDirs(searchProjectDir) {
		candidate := filepath.Join(dir, name+".html")
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		tmpl, err := template.ParseFiles(candidate)
		if err != nil {
			return Default(), fmt.Errorf("theme %q failed to parse (%s): %w; using the built-in default theme", name, candidate, err)
		}
		return &Theme{Name: name, Path: candidate, Template: tmpl}, nil
	}

	return Default(), fmt.Errorf("theme %q not found in %s; using the built-in default theme", name, strings.Join(SearchDirs(searchProjectDir), ", "))
}

// SearchDirs returns the directories searched for theme files, in order:
// the project-local <projectDir>/themes first, then <UserConfigDir>/mdoc/themes.
func SearchDirs(projectDir string) []string {
	dirs := []string{filepath.Join(projectDir, "themes")}
	if cfg, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(cfg, "mdoc", "themes"))
	}
	return dirs
}
