package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/agentskill"
	"github.com/hinkolas/mdoc/internal/browser"
	"github.com/hinkolas/mdoc/internal/paths"
)

var uninstallPurge bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove mdoc, the bundled skill, and the Chromium cache.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUninstall(uninstallPurge, stdinIsTTY && stdoutIsTTY)
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallPurge, "purge", false, "Also remove the config directory (themes) and skip all prompts")
	rootCmd.AddCommand(uninstallCmd)
}

// uninstallMode captures how runUninstall should behave: whether to ask
// interactive questions, and — when it doesn't ask — whether the config
// directory should be removed.
type uninstallMode struct {
	prompt       bool
	removeConfig bool
}

// resolveUninstallMode decides the interaction model. --purge removes
// everything (including config) with no prompts; an interactive terminal asks;
// a non-interactive terminal proceeds and keeps the config directory.
func resolveUninstallMode(purge, interactive bool) uninstallMode {
	switch {
	case purge:
		return uninstallMode{prompt: false, removeConfig: true}
	case interactive:
		return uninstallMode{prompt: true, removeConfig: false}
	default:
		return uninstallMode{prompt: false, removeConfig: false}
	}
}

func runUninstall(purge, interactive bool) error {
	mode := resolveUninstallMode(purge, interactive)

	binPath := executablePath()
	skillTargets, err := agentskill.ResolveTargets("all", "")
	if err != nil {
		return err
	}
	cacheRoot := browser.CacheRoot()
	configDir, err := paths.ConfigDir()
	if err != nil {
		return err
	}

	printBrandHeader()

	// Summary of what mdoc will touch. Present targets are highlighted; absent
	// ones are dimmed so the user sees there's nothing to do for them.
	printRow(9, "mdoc", presence(binPath))
	for _, t := range skillTargets {
		printRow(9, t.Name, presence(t.DestDir))
	}
	printRow(9, "cache", presence(cacheRoot))
	printRow(9, "config", presence(configDir))

	removeConfig := mode.removeConfig
	if mode.prompt {
		reader := bufio.NewReader(os.Stdin)
		answer, err := promptConfirm(reader, "Also remove the config directory (themes, custom CSS)?", false)
		if err != nil {
			return err
		}
		removeConfig = answer

		proceed, err := promptConfirm(reader, "Remove the items above? This cannot be undone.", false)
		if err != nil {
			return err
		}
		if !proceed {
			fmt.Printf("\n  %s uninstall cancelled — nothing was removed\n\n", red("✗"))
			return nil
		}
	}

	fmt.Println()

	// Skills.
	results, err := agentskill.Remove("all", "")
	if err != nil {
		return err
	}
	for _, res := range results {
		printRow(9, res.Target, removalStatus(res.Existed, nil))
	}

	// Chromium / mdoc cache.
	printRow(9, "cache", removalStatus(removePath(cacheRoot)))

	// Config directory (gated).
	switch {
	case removeConfig:
		printRow(9, "config", removalStatus(removePath(configDir)))
	case exists(configDir):
		printRow(9, "config", dim("kept"))
	default:
		printRow(9, "config", yellow("not found"))
	}

	// The binary last: removing the running executable is safe on macOS/Linux
	// (the inode survives until the process exits), and doing it last means a
	// failure earlier doesn't leave mdoc unusable for a retry.
	existed, rmErr := removePath(binPath)
	printRow(9, "mdoc", removalStatus(existed, rmErr))
	if rmErr != nil {
		printWarn(fmt.Sprintf("could not remove %s: %v — remove it manually", displayPath(binPath), rmErr))
	}

	fmt.Println()
	return nil
}

// executablePath returns the real path of the running mdoc binary, resolving
// any symlink. It returns "" if the path can't be determined.
func executablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		return resolved
	}
	return exe
}

// removePath deletes path (recursively) and reports whether it existed and any
// error. A missing path is not an error.
func removePath(path string) (bool, error) {
	if path == "" {
		return false, nil
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := os.RemoveAll(path); err != nil {
		return true, err
	}
	return true, nil
}

// exists reports whether path is present on disk. An empty path is "absent".
func exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// presence renders a path in the summary: the path itself when it exists,
// a dimmed "not found" otherwise.
func presence(path string) string {
	if !exists(path) {
		return dim("not found")
	}
	return displayPath(path)
}

// removalStatus renders the outcome row for one removed target.
func removalStatus(existed bool, err error) string {
	switch {
	case err != nil:
		return red("failed")
	case existed:
		return green("removed")
	default:
		return yellow("not found")
	}
}
