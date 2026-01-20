package preview

import (
	"encoding/json"
	"log"

	"github.com/hinkolas/mdoc/src/core"
)

// RenderResult is the data payload for the render_result event
type RenderResult struct {
	HTML string `json:"html"`
}

// handleRenderRequest handles the render_request event from the client
func (s *PreviewServer) handleRenderRequest(data json.RawMessage) error {
	if s.DocumentPath == "" {
		log.Println("No document path configured")
		return nil
	}

	// Open and parse the document
	doc, err := core.OpenDocument(s.DocumentPath)
	if err != nil {
		log.Printf("Failed to open document: %v", err)
		return err
	}

	// Render the document
	renderData, err := doc.Render()
	if err != nil {
		log.Printf("Render failed: %v", err)
		return err
	}

	// Send the render result back to the client
	return s.SendEventWithData("render_result", RenderResult{HTML: string(renderData.Body)})
}
