// Package theme resolves and loads HTML theme templates. Lookup order is
// project-local ./themes/<name>.html first, then the user config directory.
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

// Resolve finds a theme by name. searchProjectDir is typically the directory
// the document lives in; the function looks for <searchProjectDir>/themes/<name>.html
// first and falls back to <UserConfigDir>/mdoc/themes/<name>.html. An empty
// name yields a minimal default theme that just emits {{.Body}}.
func Resolve(name, searchProjectDir string) (*Theme, error) {
	if name == "" {
		tmpl, err := template.New("default").Parse("{{.Body}}")
		if err != nil {
			return nil, fmt.Errorf("default theme: %w", err)
		}
		return &Theme{Name: "default", Template: tmpl}, nil
	}

	for _, dir := range searchPaths(searchProjectDir) {
		candidate := filepath.Join(dir, name+".html")
		if _, err := os.Stat(candidate); err == nil {
			tmpl, err := template.ParseFiles(candidate)
			if err != nil {
				return nil, fmt.Errorf("parse theme %s: %w", candidate, err)
			}
			return &Theme{Name: name, Path: candidate, Template: tmpl}, nil
		}
	}

	return nil, fmt.Errorf("theme %q not found in %s", name, formatPaths(searchProjectDir))
}

func searchPaths(projectDir string) []string {
	paths := []string{filepath.Join(projectDir, "themes")}
	if cfg, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(cfg, "mdoc", "themes"))
	}
	return paths
}

func formatPaths(projectDir string) string {
	return strings.Join(searchPaths(projectDir), ", ")
}
