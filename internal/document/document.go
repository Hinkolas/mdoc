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
	MDoc   bool           `yaml:"mdoc"`
	Theme  string         `yaml:"theme"`
	Title  string         `yaml:"title"`
	Author string         `yaml:"author"`
	Tags   []string       `yaml:"tags"`
	Data   map[string]any `yaml:"data"`
}

// Default is applied when a file has no frontmatter or its frontmatter does not
// opt in with `mdoc: true`.
var Default = Config{
	MDoc:   true,
	Theme:  "plain",
	Title:  "Untitled",
	Author: "Anonymous",
	Tags:   []string{},
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
