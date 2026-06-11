// Package agentskill installs the bundled mdoc authoring skill into agent
// skill directories.
package agentskill

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const SkillName = "mdoc"

//go:embed mdoc
var bundled embed.FS

// InstallResult describes one installed skill target.
type InstallResult struct {
	Target    string
	ParentDir string
	DestDir   string
	Files     int
}

// Target describes one resolved install destination.
type Target struct {
	Name      string
	ParentDir string
	DestDir   string
}

// ResolveTargets expands a target name into concrete parent/destination
// directories. customParent is the parent skills directory; the skill itself is
// always installed as <customParent>/mdoc.
func ResolveTargets(name, customParent string) ([]Target, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "":
		return nil, fmt.Errorf("missing skill target")
	case "all":
		if customParent != "" {
			return nil, fmt.Errorf("--path can only be used with a single --skill target")
		}
		claude, err := ResolveTargets("claude", "")
		if err != nil {
			return nil, err
		}
		codex, err := ResolveTargets("codex", "")
		if err != nil {
			return nil, err
		}
		return append(claude, codex...), nil
	case "claude", "codex":
		parent := customParent
		if parent == "" {
			var err error
			parent, err = defaultParentDir(name)
			if err != nil {
				return nil, err
			}
		}
		return []Target{{
			Name:      name,
			ParentDir: parent,
			DestDir:   filepath.Join(parent, SkillName),
		}}, nil
	default:
		return nil, fmt.Errorf("unsupported skill target %q (use claude, codex, or all)", name)
	}
}

// Install copies the bundled mdoc skill to the requested agent target(s).
func Install(name, customParent string) ([]InstallResult, error) {
	targets, err := ResolveTargets(name, customParent)
	if err != nil {
		return nil, err
	}
	results := make([]InstallResult, 0, len(targets))
	for _, target := range targets {
		files, err := CopyTo(target.DestDir)
		if err != nil {
			return nil, fmt.Errorf("install %s skill: %w", target.Name, err)
		}
		results = append(results, InstallResult{
			Target:    target.Name,
			ParentDir: target.ParentDir,
			DestDir:   target.DestDir,
			Files:     files,
		})
	}
	return results, nil
}

// CopyTo copies the bundled mdoc skill into destDir, overwriting files that
// already exist and leaving unrelated files alone.
func CopyTo(destDir string) (int, error) {
	if destDir == "" {
		return 0, fmt.Errorf("destination directory is empty")
	}
	src, err := fs.Sub(bundled, SkillName)
	if err != nil {
		return 0, err
	}
	files := 0
	err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		dst := filepath.Join(destDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
		files++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return files, nil
}

func defaultParentDir(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch name {
	case "claude":
		return filepath.Join(home, ".claude", "skills"), nil
	case "codex":
		return filepath.Join(home, ".codex", "skills"), nil
	default:
		return "", fmt.Errorf("unsupported skill target %q", name)
	}
}
