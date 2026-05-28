package preview

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

// openInDefaultBrowser launches the system's default handler for the
// given URL — the equivalent of double-clicking it in the file manager.
// Only http/https/mailto schemes are accepted; anything else is rejected
// so the exec call can't be tricked into running an arbitrary scheme
// handler from untrusted document content.
func openInDefaultBrowser(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "mailto":
	default:
		return fmt.Errorf("disallowed scheme: %s", u.Scheme)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", raw)
	case "linux":
		cmd = exec.Command("xdg-open", raw)
	case "windows":
		// The empty "" arg gives `start` a blank window title so the
		// URL isn't misparsed as one — a long-standing Windows gotcha.
		cmd = exec.Command("cmd", "/c", "start", "", raw)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
