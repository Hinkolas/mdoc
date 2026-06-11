package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/agentskill"
	"github.com/hinkolas/mdoc/internal/browser"
)

var installRevision int
var installSkillTarget string
var installSkillPath string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up Chromium and optional agent skills.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		plan, err := buildInstallPlan(
			cmd.Flags().Changed("chromium"),
			installRevision,
			installSkillTarget,
			installSkillPath,
			args,
			stdinIsTTY && stdoutIsTTY,
		)
		if err != nil {
			return err
		}
		if plan.Interactive {
			return runInstallWizard(plan.Revision)
		}
		return runInstallDirect(plan)
	},
}

func init() {
	installCmd.Flags().IntVar(&installRevision, "chromium", -1, "Install Chromium, optionally pinned to a revision")
	installCmd.Flags().Lookup("chromium").NoOptDefVal = "-1"
	installCmd.Flags().StringVar(&installSkillTarget, "skill", "", "Install the bundled mdoc skill for claude, codex, or all")
	installCmd.Flags().StringVar(&installSkillPath, "path", "", "Parent skills directory for a single --skill target")
	rootCmd.AddCommand(installCmd)
}

type installPlan struct {
	Interactive     bool
	InstallChromium bool
	InstallSkill    bool
	Revision        int
	SkillTarget     string
	SkillPath       string
}

func buildInstallPlan(chromiumChanged bool, revision int, skillTarget, skillPath string, args []string, interactive bool) (installPlan, error) {
	if len(args) > 0 {
		if !chromiumChanged || revision != -1 {
			return installPlan{}, fmt.Errorf("unexpected argument %q", args[0])
		}
		parsed, err := strconv.Atoi(args[0])
		if err != nil {
			return installPlan{}, fmt.Errorf("invalid Chromium revision %q", args[0])
		}
		revision = parsed
		chromiumChanged = true
	}
	if skillPath != "" && skillTarget == "" {
		return installPlan{}, fmt.Errorf("--path requires --skill")
	}
	if skillTarget != "" {
		if _, err := agentskill.ResolveTargets(skillTarget, skillPath); err != nil {
			return installPlan{}, err
		}
	}

	flagsPresent := chromiumChanged || skillTarget != "" || skillPath != ""
	if !flagsPresent {
		if interactive {
			return installPlan{Interactive: true, Revision: revision}, nil
		}
		return installPlan{InstallChromium: true, Revision: revision}, nil
	}
	return installPlan{
		InstallChromium: chromiumChanged,
		InstallSkill:    skillTarget != "",
		Revision:        revision,
		SkillTarget:     skillTarget,
		SkillPath:       skillPath,
	}, nil
}

func runInstallDirect(plan installPlan) error {
	printedHeader := false
	if plan.InstallChromium {
		ui := newInstallUI()
		res, err := browser.Install(plan.Revision, ui)
		if err != nil {
			ui.abort()
			return err
		}
		ui.finish(res)
		printedHeader = true
	}
	if plan.InstallSkill {
		results, err := agentskill.Install(plan.SkillTarget, plan.SkillPath)
		if err != nil {
			return err
		}
		if !printedHeader {
			printBrandHeader()
			printedHeader = true
		}
		printSkillResults(results)
		fmt.Println()
	}
	return nil
}

func runInstallWizard(revision int) error {
	printBrandHeader()
	reader := bufio.NewReader(os.Stdin)

	status, err := browser.DetectChromium(revision)
	if err != nil {
		return err
	}
	switch {
	case status.Packaged != nil:
		printRow(10, "chromium", green("ready")+" "+dim(fmt.Sprintf("(revision %d, packaged)", status.Packaged.Revision)))
		printRow(10, "path", status.Packaged.BinPath)
		choice, err := promptChoice(reader, "Chromium is already installed. What should mdoc do?", []promptOption{
			{Value: "keep", Label: "Use cached Chromium"},
			{Value: "download", Label: "Download packaged Chromium again"},
		}, 0)
		if err != nil {
			return err
		}
		if choice == "download" {
			if _, err := installChromium(revision, true, true); err != nil {
				return err
			}
		}
	case status.System != "":
		printRow(10, "chromium", green("found")+" "+dim("(system)"))
		printRow(10, "path", status.System)
		choice, err := promptChoice(reader, "Use system Chromium or install mdoc's packaged Chromium?", []promptOption{
			{Value: "system", Label: "Use system Chromium"},
			{Value: "download", Label: "Download packaged Chromium"},
		}, 0)
		if err != nil {
			return err
		}
		if choice == "download" {
			if _, err := installChromium(revision, false, true); err != nil {
				return err
			}
		}
	default:
		choice, err := promptChoice(reader, "No Chromium was found. Download mdoc's packaged Chromium?", []promptOption{
			{Value: "yes", Label: "Download packaged Chromium"},
			{Value: "no", Label: "Skip Chromium"},
		}, 0)
		if err != nil {
			return err
		}
		if choice == "yes" {
			if _, err := installChromium(revision, false, true); err != nil {
				return err
			}
		} else {
			printRow(10, "chromium", yellow("skipped"))
		}
	}

	skillChoice, err := promptChoice(reader, "Install the bundled mdoc skill for an agent?", []promptOption{
		{Value: "none", Label: "None"},
		{Value: "claude", Label: "Claude"},
		{Value: "codex", Label: "Codex"},
		{Value: "all", Label: "Claude and Codex"},
	}, 0)
	if err != nil {
		return err
	}
	if skillChoice == "none" {
		printRow(10, "skills", yellow("skipped"))
		fmt.Println()
		return nil
	}
	results, err := agentskill.Install(skillChoice, "")
	if err != nil {
		return err
	}
	printSkillResults(results)
	fmt.Println()
	return nil
}

func installChromium(revision int, fresh bool, headerAlreadyShown bool) (*browser.InstallResult, error) {
	ui := newInstallUI()
	ui.headerShown = headerAlreadyShown
	var (
		res *browser.InstallResult
		err error
	)
	if fresh {
		res, err = browser.InstallFresh(revision, ui)
	} else {
		res, err = browser.Install(revision, ui)
	}
	if err != nil {
		ui.abort()
		return nil, err
	}
	ui.finish(res)
	return res, nil
}

func printSkillResults(results []agentskill.InstallResult) {
	for _, res := range results {
		printRow(10, res.Target, green("installed")+" "+dim(fmt.Sprintf("(%d files)", res.Files)))
		printRow(10, "path", res.DestDir)
	}
}

type promptOption struct {
	Value string
	Label string
}

func promptChoice(reader *bufio.Reader, question string, options []promptOption, defaultIndex int) (string, error) {
	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}
	fmt.Printf("\n  %s %s\n", cyan("?"), question)
	for i, opt := range options {
		label := opt.Label
		if i == defaultIndex {
			label += " " + dim("(default)")
		}
		fmt.Printf("  %s %d. %s\n", dim("│"), i+1, label)
	}
	for {
		fmt.Printf("  %s ", dim("choice:"))
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "" {
			return options[defaultIndex].Value, nil
		}
		if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(options) {
			return options[n-1].Value, nil
		}
		for _, opt := range options {
			if line == opt.Value || strings.HasPrefix(opt.Value, line) {
				return opt.Value, nil
			}
		}
		fmt.Printf("  %s enter a number from 1 to %d\n", yellow("warning:"), len(options))
	}
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
