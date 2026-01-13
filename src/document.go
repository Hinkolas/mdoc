package src

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Document struct {
	Config DocumentConfig
	Body   string
	theme  *template.Template
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

const FALLBACK_THEME = "<!doctype html><html><head><title>{{.Title}}</title></head><body>{{.Body}}</body></html>"

// const THEME_DIR = "${HOME}/.config/mdoc/themes"
const THEME_DIR = "./themes" // TODO: Remove in future releases

// TODO: Generates the pdf using headless chromium and saves it to the given path.
func (d *Document) Save(outputPath string, exportHTML bool) error {

	// 0. Render the document into clean html
	body, err := d.Render()
	if err != nil {
		return err
	}

	// Optionally export the raw HTML
	if exportHTML {
		htmlPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".html"
		err = utils.OutputFile(htmlPath, body)
		if err != nil {
			return fmt.Errorf("failed to write HTML file: %w", err)
		}
		fmt.Printf("Exported HTML to %s\n", htmlPath)
	}

	userConfig, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Failed to get user config directory:", err)
		os.Exit(1)
	}

	// Create the browser launcher instance
	chromium := launcher.NewBrowser()
	chromium.RootDir = filepath.Join(userConfig, "mdoc", "chromium")

	var binPath string
	err = chromium.Validate()
	if err != nil {
		fmt.Println("Unable to find packaged browser. Looking for local alternative...")
		var has bool = false
		binPath, has = launcher.LookPath()
		if !has {
			fmt.Println("No compatible browser found! Please run `mdoc install` to download the latest chromium snapshot.")
			os.Exit(1)
		}
	} else {
		binPath = chromium.BinPath()
	}

	fmt.Println("Initializing Browser...")

	// 2. This attempts to launch the found Chrome installation.
	u, err := launcher.New().Bin(binPath).Launch()
	if err != nil {
		fmt.Println("Failed to launch browser:", err)
		os.Exit(1)
	}

	// 3. Connect to the browser
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	// 4. Create a page (tab)
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create page: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Loading HTML into browser...")

	// Load the content into the page
	// Rod expects a string, so we convert the buffer contents.
	page.MustSetDocumentContent(body)
	page.MustWaitStable()

	fmt.Println("Generating PDF...")

	// 8. PDF Options
	// You can strictly type the config using the proto package
	paperWidth := 8.27   // A4 Width (inches)
	paperHeight := 11.69 // A4 Height (inches)
	margin := 0.0
	printBg := true
	pdfData, err := page.PDF(&proto.PagePrintToPDF{
		PaperWidth:      &paperWidth,
		PaperHeight:     &paperHeight,
		MarginTop:       &margin,
		MarginBottom:    &margin,
		MarginLeft:      &margin,
		MarginRight:     &margin,
		PrintBackground: printBg, // Important for CSS backgrounds!
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate PDF: %v\n", err)
		os.Exit(1)
	}

	// 9. Save to disk
	_ = utils.OutputFile(outputPath, pdfData)
	fmt.Printf("Done! Saved to %s\n", outputPath)

	return nil
}

type RenderData struct {
	Title  string
	Author string
	Tags   []string
	Data   map[string]any
	Body   template.HTML
}

func (d *Document) Render() (string, error) {

	var data = RenderData{
		Title:  d.Config.Title,
		Author: d.Config.Author,
		Tags:   d.Config.Tags,
		Data:   d.Config.Data,
		Body:   "",
	}

	// 1. Replace all dynamic variables in the markdown body
	var mdBuf bytes.Buffer
	tmpl, err := template.New("body").Parse(d.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse body template: %w", err)
	}

	if err := tmpl.Execute(&mdBuf, d.Config); err != nil {
		return "", fmt.Errorf("failed to execute body template: %w", err)
	}

	// Convert markdown to HTML using goldmark with GFM extensions (tables, strikethrough, etc.)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,   // Github Flavored Markdown (optional but recommended)
			mathjax.MathJax, // The Star of the Show
			extension.Footnote,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			// "Clean" HTML usually implies avoiding excessive escaping where unsafe
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var bodyBuf bytes.Buffer
	if err := md.Convert(mdBuf.Bytes(), &bodyBuf); err != nil {
		return "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	data.Body = template.HTML(bodyBuf.String())

	// 2. Insert the rendered html body and config into the theme
	var htmlBuf bytes.Buffer
	if err := d.theme.Execute(&htmlBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute theme template: %w", err)
	}

	return htmlBuf.String(), nil

}

// Creates an text object from the given input string.
// Enables special syntax like YAML front matter and manual page-break.
func ParseDocument(r io.Reader) (*Document, error) {

	var config DocumentConfig

	body, err := frontmatter.Parse(r, &config)
	if err != nil {
		return nil, err
	}

	if !config.MDoc {
		config = DEFAULT_CONFIG
	}

	// Load theme from file or fallback
	var themeDef []byte
	if config.Theme == "" {
		themeDef = []byte(FALLBACK_THEME)
	} else {
		themeFile, err := os.Open(filepath.Join(os.ExpandEnv(THEME_DIR), config.Theme+".html"))
		if err != nil {
			return nil, fmt.Errorf("failed to open theme file: %w", err)
		}
		defer themeFile.Close()

		themeDef, err = io.ReadAll(themeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read theme file: %w", err)
		}
	}

	theme, err := template.New("theme").Parse(string(themeDef))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	doc := &Document{
		Config: config,
		Body:   string(body),
		theme:  theme,
	}

	return doc, nil

}
