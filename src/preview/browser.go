package preview

import (
	"fmt"
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
		fmt.Println("Unable to find packaged browser. Looking for local alternative...")
		var has bool = false
		binPath, has = launcher.LookPath()
		if !has {
			fmt.Println("No compatible browser found! Please run `mdoc install` to download the latest chromium snapshot.")
			return nil, fmt.Errorf("no compatible browser found")
		}
	} else {
		binPath = chromium.BinPath()
	}

	fmt.Println("Initializing Browser...")

	// 2. This attempts to launch the found Chrome installation.
	u, err := launcher.New().Bin(binPath).Headless(false).Launch()
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

func (b *Browser) Close() {
	b.browser.MustClose()
}

func (b *Browser) Page() *rod.Page {
	return b.browser.MustPage()
}
