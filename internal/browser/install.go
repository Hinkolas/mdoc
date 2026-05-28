package browser

import (
	"fmt"
	"os"

	"github.com/go-rod/rod/lib/launcher"
)

// Install downloads a Chromium snapshot into <UserCacheDir>/mdoc/chromium.
// If revision is < 0, the latest revision known to go-rod is used. If a
// compatible binary is already present at the cache location, no download
// happens and its path is returned.
func Install(revision int) (string, error) {
	root := packagedRoot()
	if root == "" {
		return "", fmt.Errorf("cannot determine user cache dir")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("mkdir cache: %w", err)
	}

	br := launcher.NewBrowser()
	br.RootDir = root
	if revision >= 0 {
		br.Revision = revision
	}
	if err := br.Validate(); err == nil {
		return br.BinPath(), nil
	}
	if err := br.Download(); err != nil {
		return "", fmt.Errorf("download chromium: %w", err)
	}
	return br.BinPath(), nil
}
