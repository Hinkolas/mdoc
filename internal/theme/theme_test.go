package theme

import (
	"os"
	"path/filepath"
	"testing"
)

// minimalTheme is a parseable theme body distinguishable from the built-ins.
const minimalTheme = "<html><body>{{.Body}}</body></html>"

// setup points the user themes dir at a temp config dir and returns it.
func setup(t *testing.T) (cfgThemes string) {
	t.Helper()
	cfg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfg)
	dir := filepath.Join(cfg, "mdoc", "themes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeTheme(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(minimalTheme), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveEmptyUsesDefault(t *testing.T) {
	setup(t)
	thm, err := Resolve("", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if thm.Name != DefaultName {
		t.Errorf("Name = %q, want %q", thm.Name, DefaultName)
	}
}

func TestResolveBuiltinKeys(t *testing.T) {
	setup(t)
	for _, name := range []string{DefaultName, NoneName} {
		thm, err := Resolve(name, t.TempDir())
		if err != nil {
			t.Errorf("%s: unexpected diagnostic: %v", name, err)
		}
		if thm.Name != name {
			t.Errorf("Name = %q, want %q", thm.Name, name)
		}
	}
}

func TestResolveKeyFromUserThemesDir(t *testing.T) {
	cfgThemes := setup(t)
	writeTheme(t, filepath.Join(cfgThemes, "report.html"))

	thm, err := Resolve("report", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if thm.Path != filepath.Join(cfgThemes, "report.html") {
		t.Errorf("Path = %q, want user themes dir", thm.Path)
	}
}

// A scoped key resolves into a subdirectory of the user themes dir.
func TestResolveScopedKey(t *testing.T) {
	cfgThemes := setup(t)
	nested := filepath.Join(cfgThemes, "kilohertz", "legal", "contract.html")
	writeTheme(t, nested)

	thm, err := Resolve("kilohertz::legal::contract", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if thm.Path != nested {
		t.Errorf("Path = %q, want %q", thm.Path, nested)
	}
}

// A malformed scoped key falls back to the default with an "invalid scoped key"
// diagnostic rather than resolving somewhere unexpected.
func TestResolveScopedKeyMalformed(t *testing.T) {
	setup(t)
	thm, err := Resolve("legal::", t.TempDir())
	if err == nil {
		t.Fatal("expected diagnostic, got nil")
	}
	if thm.Name != DefaultName {
		t.Errorf("Name = %q, want fallback %q", thm.Name, DefaultName)
	}
}

// A bare key must NOT pick up a same-named theme sitting next to the document —
// that lookup is reserved for explicit paths.
func TestResolveKeyIgnoresDocLocalThemesDir(t *testing.T) {
	setup(t)
	docDir := t.TempDir()
	writeTheme(t, filepath.Join(docDir, "themes", "report.html"))

	thm, err := Resolve("report", docDir)
	if err == nil {
		t.Fatal("expected not-found diagnostic, got nil")
	}
	if thm.Name != DefaultName {
		t.Errorf("Name = %q, want fallback %q", thm.Name, DefaultName)
	}
}

func TestResolveRelativePathFromDocDir(t *testing.T) {
	setup(t)
	docDir := t.TempDir()
	writeTheme(t, filepath.Join(docDir, "themes", "thesis.html"))

	thm, err := Resolve("./themes/thesis.html", docDir)
	if err != nil {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if thm.Path != filepath.Join(docDir, "themes", "thesis.html") {
		t.Errorf("Path = %q, want doc-local themes file", thm.Path)
	}
}

func TestResolveParentRelativePath(t *testing.T) {
	setup(t)
	base := t.TempDir()
	writeTheme(t, filepath.Join(base, "thesis.html"))
	docDir := filepath.Join(base, "sub")
	if err := os.MkdirAll(docDir, 0o755); err != nil {
		t.Fatal(err)
	}

	thm, err := Resolve("../thesis.html", docDir)
	if err != nil {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if thm.Path != filepath.Join(base, "thesis.html") {
		t.Errorf("Path = %q, want %q", thm.Path, filepath.Join(base, "thesis.html"))
	}
}

func TestResolveAbsolutePath(t *testing.T) {
	setup(t)
	abs := filepath.Join(t.TempDir(), "dev.html")
	writeTheme(t, abs)

	thm, err := Resolve(abs, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if thm.Path != abs {
		t.Errorf("Path = %q, want %q", thm.Path, abs)
	}
}

func TestResolveMissingPathFallsBack(t *testing.T) {
	setup(t)
	thm, err := Resolve("./themes/nope.html", t.TempDir())
	if err == nil {
		t.Fatal("expected not-found diagnostic, got nil")
	}
	fb, ok := err.(*Fallback)
	if !ok {
		t.Fatalf("error type = %T, want *Fallback", err)
	}
	if fb.Reason != "not found" {
		t.Errorf("Reason = %q, want %q", fb.Reason, "not found")
	}
	if thm.Name != DefaultName {
		t.Errorf("Name = %q, want fallback %q", thm.Name, DefaultName)
	}
}
