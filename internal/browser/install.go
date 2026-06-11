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

// ChromiumStatus reports what mdoc can currently use before doing any
// downloads.
type ChromiumStatus struct {
	Packaged *InstallResult
	System   string
}

// DetectChromium checks the packaged cache and system PATH without changing
// either. If revision is < 0, the latest revision known to go-rod is used for
// packaged-cache validation.
func DetectChromium(revision int) (*ChromiumStatus, error) {
	packaged, err := Packaged(revision)
	if err != nil {
		return nil, err
	}
	sys, _ := launcher.LookPath()
	return &ChromiumStatus{Packaged: packaged, System: sys}, nil
}

// Packaged checks whether the managed Chromium snapshot is already installed.
// It returns nil, nil when the cache exists but no compatible binary is present.
func Packaged(revision int) (*InstallResult, error) {
	root := packagedRoot()
	if root == "" {
		return nil, fmt.Errorf("cannot determine user cache dir")
	}
	br := launcher.NewBrowser()
	br.RootDir = root
	if revision >= 0 {
		br.Revision = revision
	}
	if err := br.Validate(); err != nil {
		return nil, nil
	}
	return &InstallResult{
		BinPath:    br.BinPath(),
		Revision:   br.Revision,
		Downloaded: false,
	}, nil
}

// Install downloads a Chromium snapshot into <UserCacheDir>/mdoc/chromium.
// If revision is < 0, the latest revision known to go-rod is used. When a
// compatible binary is already present nothing is downloaded.
//
// If logger is non-nil it receives go-rod's fetchup events (Download:,
// Progress:, Downloaded:, Unzip:) for the CLI to render its own UI.
func Install(revision int, logger utils.Logger) (*InstallResult, error) {
	return install(revision, logger, false)
}

// InstallFresh downloads a Chromium snapshot even when the managed cache
// already contains a compatible binary.
func InstallFresh(revision int, logger utils.Logger) (*InstallResult, error) {
	return install(revision, logger, true)
}

func install(revision int, logger utils.Logger, force bool) (*InstallResult, error) {
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
	if !force {
		if err := br.Validate(); err == nil {
			return &InstallResult{
				BinPath:    br.BinPath(),
				Revision:   br.Revision,
				Downloaded: false,
			}, nil
		}
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
