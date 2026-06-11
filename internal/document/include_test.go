package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write creates a file under dir (creating parent dirs) and returns its path.
func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolveIncludesBasic(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "chapter1.md", "# Chapter One\n\nFirst chapter.")
	root := "Intro.\n\n:::include chapter1.md\n\nOutro."

	got, included, err := resolveIncludes(root, dir, []string{filepath.Join(dir, "root.md")})
	if err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"Intro.", "# Chapter One", "First chapter.", "Outro."} {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in:\n%s", sub, got)
		}
	}
	if strings.Contains(got, ":::include") {
		t.Errorf("include directive left in output:\n%s", got)
	}
	if len(included) != 1 || included[0] != filepath.Join(dir, "chapter1.md") {
		t.Errorf("includes = %v, want [chapter1.md]", included)
	}
}

func TestResolveIncludesNested(t *testing.T) {
	dir := t.TempDir()
	// part1/index.md includes a sibling chapter by a path relative to itself.
	write(t, dir, "part1/index.md", "# Part One\n\n:::include chapter.md")
	write(t, dir, "part1/chapter.md", "## Deep chapter")
	root := ":::include part1/index.md"

	got, included, err := resolveIncludes(root, dir, []string{filepath.Join(dir, "root.md")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "# Part One") || !strings.Contains(got, "## Deep chapter") {
		t.Errorf("nested content missing:\n%s", got)
	}
	want := []string{filepath.Join(dir, "part1/index.md"), filepath.Join(dir, "part1/chapter.md")}
	if len(included) != 2 || included[0] != want[0] || included[1] != want[1] {
		t.Errorf("includes = %v, want %v", included, want)
	}
}

func TestResolveIncludesStripsFrontmatter(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "chapter.md", "---\nmdoc: true\ntitle: Standalone\ntheme: other\n---\n# Body Heading\n\nText.")
	root := ":::include chapter.md"

	got, _, err := resolveIncludes(root, dir, []string{filepath.Join(dir, "root.md")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "# Body Heading") {
		t.Errorf("body missing:\n%s", got)
	}
	for _, leaked := range []string{"mdoc: true", "title: Standalone", "theme: other", "---"} {
		if strings.Contains(got, leaked) {
			t.Errorf("frontmatter leaked (%q) into:\n%s", leaked, got)
		}
	}
}

func TestResolveIncludesMissingFile(t *testing.T) {
	dir := t.TempDir()
	root := ":::include nope.md"
	_, _, err := resolveIncludes(root, dir, []string{filepath.Join(dir, "root.md")})
	if err == nil {
		t.Fatal("expected an error for a missing include")
	}
	if !strings.Contains(err.Error(), "nope.md") {
		t.Errorf("error should name the missing file, got: %v", err)
	}
}

func TestResolveIncludesCycle(t *testing.T) {
	dir := t.TempDir()
	// a.md -> b.md -> a.md
	write(t, dir, "a.md", ":::include b.md")
	write(t, dir, "b.md", ":::include a.md")
	rootPath := filepath.Join(dir, "root.md")
	root := ":::include a.md"

	_, _, err := resolveIncludes(root, dir, []string{rootPath})
	if err == nil {
		t.Fatal("expected a cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected a cycle error, got: %v", err)
	}
}

func TestResolveIncludesIgnoresFencedCode(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "chapter.md", "real content")
	root := strings.Join([]string{
		"```",
		":::include chapter.md",
		"```",
		"",
		":::include chapter.md",
	}, "\n")

	got, included, err := resolveIncludes(root, dir, []string{filepath.Join(dir, "root.md")})
	if err != nil {
		t.Fatal(err)
	}
	// The fenced occurrence stays literal; only the real one is spliced.
	if !strings.Contains(got, ":::include chapter.md") {
		t.Errorf("fenced include directive should be preserved verbatim:\n%s", got)
	}
	if !strings.Contains(got, "real content") {
		t.Errorf("real include not spliced:\n%s", got)
	}
	if len(included) != 1 {
		t.Errorf("expected exactly one resolved include, got %v", included)
	}
}

func TestResolveIncludesNoIncludes(t *testing.T) {
	dir := t.TempDir()
	root := "# Just a heading\n\nNo includes here."
	got, included, err := resolveIncludes(root, dir, []string{filepath.Join(dir, "root.md")})
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Errorf("body should be unchanged, got:\n%s", got)
	}
	if included != nil {
		t.Errorf("expected no includes, got %v", included)
	}
}
