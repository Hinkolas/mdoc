package agentskill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetsDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name string
		want string
	}{
		{"claude", filepath.Join(home, ".claude", "skills", "mdoc")},
		{"codex", filepath.Join(home, ".codex", "skills", "mdoc")},
	}

	for _, tt := range tests {
		targets, err := ResolveTargets(tt.name, "")
		if err != nil {
			t.Fatalf("ResolveTargets(%q): %v", tt.name, err)
		}
		if len(targets) != 1 {
			t.Fatalf("ResolveTargets(%q) returned %d targets, want 1", tt.name, len(targets))
		}
		if targets[0].DestDir != tt.want {
			t.Fatalf("ResolveTargets(%q) dest = %q, want %q", tt.name, targets[0].DestDir, tt.want)
		}
	}
}

func TestResolveTargetsAll(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	targets, err := ResolveTargets("all", "")
	if err != nil {
		t.Fatalf("ResolveTargets(all): %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("ResolveTargets(all) returned %d targets, want 2", len(targets))
	}
	if targets[0].Name != "claude" || targets[1].Name != "codex" {
		t.Fatalf("ResolveTargets(all) names = %q, %q; want claude, codex", targets[0].Name, targets[1].Name)
	}
}

func TestResolveTargetsCustomParent(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "skills")
	targets, err := ResolveTargets("claude", parent)
	if err != nil {
		t.Fatalf("ResolveTargets custom parent: %v", err)
	}
	if got, want := targets[0].DestDir, filepath.Join(parent, "mdoc"); got != want {
		t.Fatalf("custom dest = %q, want %q", got, want)
	}
}

func TestResolveTargetsInvalid(t *testing.T) {
	if _, err := ResolveTargets("unknown", ""); err == nil {
		t.Fatal("ResolveTargets unknown target succeeded, want error")
	}
	if _, err := ResolveTargets("all", t.TempDir()); err == nil {
		t.Fatal("ResolveTargets all with custom path succeeded, want error")
	}
	if _, err := ResolveTargets("", ""); err == nil {
		t.Fatal("ResolveTargets empty target succeeded, want error")
	}
}

func TestCopyToCopiesBundledSkill(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "mdoc")
	if err := os.MkdirAll(filepath.Join(dest, "examples"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := CopyTo(dest)
	if err != nil {
		t.Fatalf("CopyTo: %v", err)
	}
	if files < 8 {
		t.Fatalf("CopyTo copied %d files, want at least 8", files)
	}
	for _, rel := range []string{
		"SKILL.md",
		"frontmatter.md",
		"syntax.md",
		"themes.md",
		"cli.md",
		filepath.Join("examples", "document.md"),
		filepath.Join("examples", "plain.html"),
		filepath.Join("examples", "assets", "pipeline.svg"),
	} {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			t.Fatalf("expected copied file %s: %v", rel, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "stale" {
		t.Fatal("CopyTo did not overwrite stale SKILL.md")
	}
}
