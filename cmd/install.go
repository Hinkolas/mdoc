package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/browser"
)

var installRevision int

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download a compatible Chromium snapshot into the user cache directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui := newInstallUI()
		res, err := browser.Install(installRevision, ui)
		if err != nil {
			ui.abort()
			return err
		}
		ui.finish(res)
		return nil
	},
}

func init() {
	installCmd.Flags().IntVar(&installRevision, "chromium", -1, "Specific Chromium revision to install (default: latest known)")
	rootCmd.AddCommand(installCmd)
}

// installUI translates go-rod's fetchup events (Download:, Progress:,
// Downloaded:, Unzip:) into a small in-place progress bar that fits the
// rest of mdoc's styling. Implements utils.Logger.
type installUI struct {
	phase       string // "downloading" | "extracting" | ""
	headerShown bool
	barWidth    int
	lastBar     bool // last printed line was a progress bar
}

func newInstallUI() *installUI { return &installUI{barWidth: 32} }

// Println is the hook go-rod's launcher invokes for fetch events. Each
// call is a logical "log line"; we route them to phase/progress handlers.
func (u *installUI) Println(vs ...interface{}) {
	if len(vs) == 0 {
		return
	}
	// fetchup's event constants are a typed string (fetchup.Event), not
	// plain string, so use Sprint instead of a type assertion.
	event := strings.TrimSuffix(fmt.Sprint(vs[0]), ":")
	var arg string
	if len(vs) > 1 {
		arg = fmt.Sprint(vs[1])
	}
	switch event {
	case "Download":
		u.ensureHeader()
		u.startPhase("downloading chromium")
	case "Unzip":
		u.startPhase("extracting")
	case "Progress":
		u.renderProgress(arg)
	case "Downloaded":
		u.completePhase()
	}
}

func (u *installUI) ensureHeader() {
	if u.headerShown {
		return
	}
	printBrandHeader()
	u.headerShown = true
}

func (u *installUI) startPhase(name string) {
	// Cap the previous phase's bar at 100% so it doesn't read as
	// "abandoned mid-flight" in the scrollback. Only emit the bar +
	// terminating newline when we actually drew a bar — otherwise we
	// leave a confusing blank line in non-TTY output.
	if u.lastBar {
		u.renderProgress("100")
		fmt.Println()
		u.lastBar = false
	}
	u.phase = name
	fmt.Printf("  %s  %s\n", cyan("➜"), name)
	u.renderProgress("0")
}

// renderProgress overwrites the current bar line in place when on a TTY;
// in a pipe we omit the bar entirely since \r looks like garbage in logs.
func (u *installUI) renderProgress(pctStr string) {
	if !stdoutIsTTY {
		return
	}
	pct, err := strconv.Atoi(strings.TrimSuffix(pctStr, "%"))
	if err != nil {
		return
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := pct * u.barWidth / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", u.barWidth-filled)
	fmt.Printf("\r  %s %s %s%%", dim("│"), bar, dim(fmt.Sprintf("%3d", pct)))
	u.lastBar = true
}

func (u *installUI) completePhase() {
	if u.lastBar {
		u.renderProgress("100")
		fmt.Println()
		u.lastBar = false
	}
	u.phase = ""
}

// abort cleans the cursor up so an error message doesn't land on the same
// line as a half-drawn progress bar.
func (u *installUI) abort() {
	if u.lastBar {
		fmt.Println()
		u.lastBar = false
	}
}

func (u *installUI) finish(res *browser.InstallResult) {
	u.completePhase()
	u.ensureHeader()

	status := green("installed") + " " + dim(fmt.Sprintf("(revision %d)", res.Revision))
	if !res.Downloaded {
		status = green("ready") + " " + dim(fmt.Sprintf("(revision %d, cached)", res.Revision))
	}
	printRow(9, "chromium", status)
	printRow(9, "path", res.BinPath)
	fmt.Println()
}
