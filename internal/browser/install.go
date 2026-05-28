package browser

import (
	"fmt"
	"os"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/utils"
)

// InstallResult describes the outcome of an Install call.
type InstallResult struct {
	// BinPath is the absolute path of the Chromium executable.
	BinPath string
	// Revision is the Chromium revision that ended up installed.
	Revision int
	// Downloaded is true when this call actually fetched and unpacked a
	// new Chromium; false when a compatible one was already in the cache.
	Downloaded bool
}

// Install downloads a Chromium snapshot into <UserCacheDir>/mdoc/chromium.
// If revision is < 0, the latest revision known to go-rod is used. When a
// compatible binary is already present nothing is downloaded.
//
// If logger is non-nil it receives go-rod's fetchup events (Download:,
// Progress:, Downloaded:, Unzip:) for the CLI to render its own UI.
func Install(revision int, logger utils.Logger) (*InstallResult, error) {
	root := packagedRoot()
	if root == "" {
		return nil, fmt.Errorf("cannot determine user cache dir")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir cache: %w", err)
	}

	br := launcher.NewBrowser()
	br.RootDir = root
	if revision >= 0 {
		br.Revision = revision
	}
	if err := br.Validate(); err == nil {
		return &InstallResult{
			BinPath:    br.BinPath(),
			Revision:   br.Revision,
			Downloaded: false,
		}, nil
	}

	// Only swap the logger right before we actually start downloading —
	// Validate() shouldn't print anything, but this keeps the contract
	// crisp: events only fire during the download/unzip phases.
	if logger != nil {
		br.Logger = logger
	}
	if err := br.Download(); err != nil {
		return nil, fmt.Errorf("download chromium: %w", err)
	}
	return &InstallResult{
		BinPath:    br.BinPath(),
		Revision:   br.Revision,
		Downloaded: true,
	}, nil
}
