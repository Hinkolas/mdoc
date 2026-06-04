// Package document parses a markdown source file with YAML frontmatter into a
// Document value. The Document only carries the data parsed off disk; rendering
// lives in internal/render and theme resolution in internal/theme.
package document

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/frontmatter"
)

// Config is the YAML frontmatter shape.
type Config struct {
	MDoc       bool           `yaml:"mdoc"`
	Theme      string         `yaml:"theme"`
	Title      string         `yaml:"title"`
	Author     string         `yaml:"author"`
	Tags       []string       `yaml:"tags"`
	Page       Page           `yaml:"page"`
	Data       map[string]any `yaml:"data"`
	References []Reference    `yaml:"references"`
	Numbering  Numbering      `yaml:"numbering"`
}

// Reference is one bibliography entry. Cited from the body with `[@<key>]` and
// listed by a `:::bibliography` directive. If Text is set it is used verbatim
// (the raw escape-hatch); otherwise the structured fields are assembled by the
// renderer. Both `key` and `id` name the citation key.
type Reference struct {
	Key       string `yaml:"key"`
	ID        string `yaml:"id"`
	Author    string `yaml:"author"`
	Title     string `yaml:"title"`
	Year      string `yaml:"year"`
	Publisher string `yaml:"publisher"`
	Edition   string `yaml:"edition"`
	ISBN      string `yaml:"isbn"`
	URL       string `yaml:"url"`
	Text      string `yaml:"text"`
}

// CiteKey is the key a `[@…]` citation matches against (`key`, or `id` as an
// alias).
func (r Reference) CiteKey() string {
	if r.Key != "" {
		return r.Key
	}
	return r.ID
}

// Numbering configures automatic heading numbering. It is off by default so
// ordinary documents don't get "1", "1.1" prefixes; thesis/report documents
// opt in with `numbering: {enabled: true}`. A `:::toc` works either way (entries
// just carry no number when numbering is off).
type Numbering struct {
	Enabled bool `yaml:"enabled"`
}

// Page mirrors the relevant parts of CSS @page. Both fields are passed
// through verbatim into the theme's @page rule, so anything CSS accepts
// (named sizes like "A4" / "Letter", explicit "210mm 297mm", "A4 landscape",
// the four-value margin shorthand, etc.) is valid. Themes provide the
// fallback when a field is empty.
type Page struct {
	Size   string `yaml:"size"`
	Margin string `yaml:"margin"`
}

// Default is applied when a file has no frontmatter or its frontmatter does not
// opt in with `mdoc: true`.
var Default = Config{
	MDoc:   true,
	Theme:  "", // empty -> built-in default theme ("system"); see internal/theme.Resolve
	Title:  "Untitled",
	Author: "Anonymous",
	Tags:   []string{},
	Page:   Page{},
	Data:   map[string]any{},
}

// Document is a parsed markdown source file.
type Document struct {
	Config Config
	Body   string

	// Path is the absolute path to the source file.
	Path string
	// Dir is the absolute directory containing the source file. Relative
	// references inside the document (images, includes) resolve from here.
	Dir string
}

// Open reads and parses a markdown file.
func Open(path string) (*Document, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open document: %w", err)
	}
	defer f.Close()

	var cfg Config
	body, err := frontmatter.Parse(f, &cfg)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	if !cfg.MDoc {
		cfg = Default
	}

	return &Document{
		Config: cfg,
		Body:   string(body),
		Path:   abs,
		Dir:    filepath.Dir(abs),
	}, nil
}
