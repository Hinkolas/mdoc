// Package browser wraps go-rod with two flavors used by mdoc:
//
//   - Headless: a one-shot headless instance for the print pipeline.
//   - AppMode: a chromeless app window pointing at a URL, used by `mdoc open`.
//
// Both share the same chromium discovery logic (packaged in the user cache
// directory, with a system Chromium as fallback).
package browser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Browser is a thin handle around a go-rod browser + initial page.
type Browser struct {
	browser *rod.Browser
	page    *rod.Page
}

// Page returns the initial page/tab created with the browser.
func (b *Browser) Page() *rod.Page { return b.page }

// RodBrowser exposes the underlying go-rod browser for callers that need
// to wait on browser-level events (e.g. detecting the user closing an
// app-mode window).
func (b *Browser) RodBrowser() *rod.Browser { return b.browser }

// Close shuts down the browser process.
func (b *Browser) Close() {
	if b.browser != nil {
		_ = b.browser.Close()
	}
}

// Headless starts a headless Chromium with a single blank page.
func Headless() (*Browser, error) {
	binPath, err := ResolveBinary()
	if err != nil {
		return nil, err
	}
	l := launcher.New().Bin(binPath).Headless(true)
	ctrlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch chromium: %w", err)
	}
	br := rod.New().ControlURL(ctrlURL)
	if err := br.Connect(); err != nil {
		return nil, fmt.Errorf("connect to chromium: %w", err)
	}
	page, err := br.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		_ = br.Close()
		return nil, fmt.Errorf("create initial page: %w", err)
	}
	return &Browser{browser: br, page: page}, nil
}

// AppMode starts Chromium in chromeless --app=<url> mode (no tab strip,
// no URL bar) so the preview window looks like a desktop app.
func AppMode(url string) (*Browser, error) {
	if url == "" {
		return nil, fmt.Errorf("AppMode requires a URL")
	}
	binPath, err := ResolveBinary()
	if err != nil {
		return nil, err
	}
	// NewAppMode is the go-rod preset that knows the magic flag dance:
	// it deletes "no-startup-window" (which the plain launcher sets by
	// default and which suppresses the --app window from ever opening)
	// and disables the automation banner.
	l := launcher.NewAppMode(url).Bin(binPath)
	ctrlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch chromium: %w", err)
	}
	br := rod.New().ControlURL(ctrlURL)
	if err := br.Connect(); err != nil {
		return nil, fmt.Errorf("connect to chromium: %w", err)
	}
	// In app mode chromium has already opened the window pointed at the
	// URL; don't create a second page (that would surface as a plain
	// about:blank window alongside the app one).
	return &Browser{browser: br}, nil
}

// ResolveBinary returns the chromium binary path: prefers the packaged copy
// under <UserCacheDir>/mdoc/chromium, falls back to a system Chromium found
// via launcher.LookPath().
func ResolveBinary() (string, error) {
	cache := packagedRoot()
	if cache != "" {
		br := launcher.NewBrowser()
		br.RootDir = cache
		if err := br.Validate(); err == nil {
			return br.BinPath(), nil
		}
	}
	if sys, ok := launcher.LookPath(); ok {
		return sys, nil
	}
	return "", fmt.Errorf("no chromium found; run `mdoc install` to download one")
}

// packagedRoot is <UserCacheDir>/mdoc/chromium, or "" if the cache dir
// isn't available on this system.
func packagedRoot() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cacheDir, "mdoc", "chromium")
}
