package core

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type RenderData struct {
	Title  string
	Author string
	Tags   []string
	Data   map[string]any
	Body   template.HTML
	System SystemData
}

func (d *Document) RenderData() (*RenderData, error) {

	now := time.Now()
	return &RenderData{
		Title:  d.Config.Title,
		Author: d.Config.Author,
		Tags:   d.Config.Tags,
		Data:   d.Config.Data,
		Body:   "",
		System: SystemData{
			Date:    now.Format("02 January 2006"),
			Time:    now.Format("15:04:05"),
			Version: "unknown", // TODO: Replace with actual mdoc version
		},
	}, nil

}

func (d *Document) Render() (*RenderData, error) {

	var data, err = d.RenderData()
	if err != nil {
		return nil, fmt.Errorf("failed to collect render data: %w", err)
	}

	// 1. Replace all dynamic variables in the markdown body
	var mdBuf bytes.Buffer
	tmpl, err := template.New("source").Parse(d.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body template: %w", err)
	}

	if err := tmpl.Execute(&mdBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute body template: %w", err)
	}

	// Convert markdown to HTML using goldmark with GFM extensions (tables, strikethrough, etc.)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,      // Enables support for Github Flavored Markdown
			mathjax.MathJax,    // Enables support for LaTeX math equations
			extension.Footnote, // Enables support for footnotes
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Adds unique IDs to headings for easier linking
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // TODO: Explain why this is needed
			html.WithXHTML(),     // TODO: Explain why this is needed
		),
	)

	var bodyBuf bytes.Buffer
	if err := md.Convert(mdBuf.Bytes(), &bodyBuf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	data.Body = template.HTML(bodyBuf.String())

	// 2. Insert the rendered html body and config into the theme
	theme, err := loadThemeTemplate(d.ThemePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load theme template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := theme.Execute(&htmlBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	data.Body = template.HTML(htmlBuf.String())

	return data, nil

}

func loadThemeTemplate(path string) (*template.Template, error) {
	if path == "" {
		fmt.Println("No theme path provided, using default theme")
		return template.New("theme").Parse("{{.Body}}")
	} else {
		return template.ParseFiles(path)
	}
}
