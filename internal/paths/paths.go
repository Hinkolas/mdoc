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
	"os"
	"path/filepath"
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
