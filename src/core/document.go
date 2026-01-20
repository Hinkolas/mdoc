package core

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/frontmatter"
)

type Document struct {
	Config       DocumentConfig
	Body         string
	ThemePath    string
	DocumentPath string
}

type DocumentConfig struct {
	MDoc   bool           `yaml:"mdoc"`
	Theme  string         `yaml:"theme"`
	Title  string         `yaml:"title"`
	Author string         `yaml:"author"`
	Tags   []string       `yaml:"tags"`
	Data   map[string]any `yaml:"data"`
}

var DEFAULT_CONFIG = DocumentConfig{
	MDoc:   true,
	Theme:  "plain",
	Title:  "Untitled",
	Author: "Anonymous",
	Tags:   []string{},
	Data:   map[string]any{},
}

// const THEME_DIR = "${HOME}/.config/mdoc/themes"
const THEME_DIR = "./themes" // TODO: Remove in future releases

func NewDocument() (*Document, error) {
	return &Document{
		Config:       DEFAULT_CONFIG,
		Body:         "",
		ThemePath:    "",
		DocumentPath: "",
	}, nil
}

func OpenDocument(path string) (*Document, error) {

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Parse markdown
	document, err := parseMarkdownSource(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing document: %w", err)
	}

	return document, nil

}

// Parses the input to produce a Document object.
// Handles YAML front matter for configuration and supports custom features such as manual page breaks (TODO).
func parseMarkdownSource(r io.Reader) (*Document, error) {

	var config DocumentConfig

	// Parse the YAML front matter config
	body, err := frontmatter.Parse(r, &config)
	if err != nil {
		return nil, err
	}

	// If the document is not a mdoc document, use the default config
	if !config.MDoc {
		config = DEFAULT_CONFIG
	}

	var themePath string
	if config.Theme != "" {
		themePath = filepath.Join(os.ExpandEnv(THEME_DIR), config.Theme+".html")

		// Check if the theme file exists
		if _, err := os.Stat(themePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("theme file not found: %s", themePath)
		}
	}

	// Create the document
	doc := &Document{
		Config:    config,
		Body:      string(body),
		ThemePath: themePath,
	}

	return doc, nil

}
