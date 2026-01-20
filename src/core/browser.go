package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Browser struct {
	browser *rod.Browser
	page    *rod.Page
}

func (b *Browser) Close() {
	b.browser.MustClose()
}

func (b *Browser) GeneratePDF(body string) (io.Reader, error) {

	// Load the content into the page
	// Rod expects a string, so we convert the buffer contents.
	b.page.MustSetDocumentContent(body)
	b.page.MustWaitStable()

	// 8. PDF Options
	// You can strictly type the config using the proto package
	paperWidth := 8.27   // A4 Width (inches)
	paperHeight := 11.69 // A4 Height (inches)
	margin := 0.0
	printBg := true
	pdfData, err := b.page.PDF(&proto.PagePrintToPDF{
		PaperWidth:      &paperWidth,
		PaperHeight:     &paperHeight,
		MarginTop:       &margin,
		MarginBottom:    &margin,
		MarginLeft:      &margin,
		MarginRight:     &margin,
		PrintBackground: printBg, // Important for CSS backgrounds!
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfData, nil

}

func StartBrowser() (*Browser, error) {

	userConfig, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}

	// Create the browser launcher instance
	chromium := launcher.NewBrowser()
	chromium.RootDir = filepath.Join(userConfig, "mdoc", "chromium")

	var binPath string
	err = chromium.Validate()
	if err != nil {
		var has bool = false
		binPath, has = launcher.LookPath()
		if !has {
			return nil, fmt.Errorf("no compatible browser found")
		}
	} else {
		binPath = chromium.BinPath()
	}

	// 2. This attempts to launch the found Chrome installation.
	u, err := launcher.New().Bin(binPath).Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// 3. Connect to the browser
	browser := rod.New().ControlURL(u).MustConnect()

	// 4. Create a page (tab)
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return &Browser{
		browser: browser,
		page:    page,
	}, nil

}
