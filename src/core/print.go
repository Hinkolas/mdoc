package core

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"

	_ "embed"
)

//go:embed print.html
var printTemplateDef string

// TODO: Generates the pdf using headless chromium and saves it to the given path.
func (d *Document) Print(outputPath string) error {

	// 0. Render the document into clean html
	data, err := d.Render()
	if err != nil {
		return err
	}

	// 2. Insert the rendered html body and config into the theme
	printTemplate, err := template.New("print").Parse(printTemplateDef)
	if err != nil {
		return fmt.Errorf("failed to parse print template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := printTemplate.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("failed to execute print template: %w", err)
	}

	os.WriteFile("debug.html", htmlBuf.Bytes(), 0644)

	browser, err := StartBrowser()
	if err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}
	defer browser.Close()

	pdfData, err := browser.GeneratePDF(htmlBuf.String())
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// 9. Save to disk
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outputFile.Close()

	// This "pipes" the reader directly to the file writer
	_, err = io.Copy(outputFile, pdfData)
	if err != nil {
		return fmt.Errorf("failed to pipe PDF data to file: %w", err)
	}

	fmt.Printf("Done! Saved to %s\n", outputPath)

	return nil

}
