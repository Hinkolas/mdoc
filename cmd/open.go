package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-rod/rod/lib/proto"
	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/browser"
	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/preview"
)

var openPort int

var openCmd = &cobra.Command{
	Use:   "open <file>",
	Short: "Open a live preview of a markdown document in a chromeless chromium window.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docPath := args[0]

		// Parse once up front so we fail fast on bad frontmatter / missing
		// theme rather than after opening a browser window.
		doc, err := document.Open(docPath)
		if err != nil {
			return err
		}

		srv := preview.New(doc.Path, Version)
		if err := srv.Start(openPort); err != nil {
			return err
		}
		defer srv.Shutdown()

		themePath, _ := srv.CurrentThemePath()
		watcher, err := preview.NewWatcher(func() {
			if err := srv.PushReload(); err != nil {
				fmt.Fprintln(os.Stderr, "reload:", err)
			}
		}, doc.Path, themePath)
		if err != nil {
			return err
		}
		defer watcher.Close()
		go watcher.Run()

		printStartupBanner(Version, srv.URL(), doc.Path, doc.Config.Theme)

		br, err := browser.AppMode(srv.URL())
		if err != nil {
			return err
		}
		defer br.Close()

		// Block until either the user closes the window or sends a signal.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		// When the user closes the --app window, Chromium destroys the
		// target. That's our cue to exit cleanly.
		windowClosed := make(chan struct{})
		go func() {
			defer close(windowClosed)
			wait := br.RodBrowser().WaitEvent(&proto.TargetTargetDestroyed{})
			wait()
		}()

		select {
		case <-sig:
			fmt.Println()
		case <-windowClosed:
		}
		return nil
	},
}

func init() {
	openCmd.Flags().IntVarP(&openPort, "port", "p", 7768, "Preview server port (0 = pick a free port)")
	rootCmd.AddCommand(openCmd)
}

// printStartupBanner writes the small Vite-style block that introduces the
// preview session — version, URL, document, theme. Colors are applied only
// when stdout is a terminal so piping to a file stays clean.
func printStartupBanner(version, url, docPath, themeName string) {
	isTTY := false
	if fi, err := os.Stdout.Stat(); err == nil {
		isTTY = (fi.Mode() & os.ModeCharDevice) != 0
	}
	style := func(code, text string) string {
		if !isTTY {
			return text
		}
		return "\033[" + code + "m" + text + "\033[0m"
	}
	bold := func(s string) string { return style("1", s) }
	dim := func(s string) string { return style("2", s) }
	cyan := func(s string) string { return style("36", s) }
	underline := func(s string) string { return style("4;36", s) }

	// Show paths relative to the user's cwd when possible — shorter and
	// usually what the user typed.
	display := docPath
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, docPath); err == nil && !strings.HasPrefix(rel, "..") {
			display = rel
		}
	}

	const labelWidth = 10
	row := func(label, value string) {
		pad := strings.Repeat(" ", labelWidth-len(label))
		fmt.Printf("  %s  %s%s%s\n", cyan("➜"), dim(label), pad, value)
	}

	fmt.Println()
	fmt.Printf("  %s  %s\n", bold("mdoc"), dim("v"+version))
	fmt.Println()
	row("preview", underline(url))
	row("document", display)
	row("theme", themeName)
	fmt.Println()
	fmt.Printf("  %s\n\n", dim("press ctrl+c to stop"))
}
