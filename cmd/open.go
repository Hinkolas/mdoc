package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-rod/rod/lib/proto"
	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/browser"
	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/preview"
	"github.com/hinkolas/mdoc/internal/theme"
)

var openPort int

var openCmd = &cobra.Command{
	Use:   "open <file>",
	Short: "Open a live preview of a markdown document in a chromeless chromium window.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docPath := args[0]

		// Parse once up front so we fail fast on bad frontmatter rather than
		// after opening a browser window. A missing/broken theme is NOT fatal —
		// Resolve falls back to the built-in default and reports a warning.
		doc, err := document.Open(docPath)
		if err != nil {
			return err
		}
		thm, twarn := theme.Resolve(doc.Config.Theme, doc.Dir)

		srv := preview.New(doc.Path, Version)
		if err := srv.Start(openPort); err != nil {
			return err
		}
		defer srv.Shutdown()

		// Watch the document and the theme search directories — so a theme
		// being created or switched is noticed — plus the active theme file,
		// which is re-pointed on every change since the active theme can move
		// around the session. The watcher is referenced from its own callback,
		// so declare it first and assign before Run starts firing.
		watchPaths := append([]string{doc.Path}, theme.SearchDirs(doc.Dir)...)
		var watcher *preview.Watcher
		watcher, err = preview.NewWatcher(func() {
			watcher.WatchTheme(srv.CurrentThemePath())
			if err := srv.PushReload(); err != nil {
				fmt.Fprintln(os.Stderr, "reload:", err)
			}
		}, watchPaths...)
		if err != nil {
			return err
		}
		defer watcher.Close()
		watcher.WatchTheme(thm.Path)
		go watcher.Run()

		printStartupBanner(Version, srv.URL(), doc.Path)
		if twarn != nil {
			printWarn(twarn.Error())
		}

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
// preview session — version, URL, document. The theme is deliberately not
// shown: it can be changed live during the session, so a fixed banner value
// would go stale (and theme problems are reported as warnings instead).
func printStartupBanner(_, url, docPath string) {
	display := displayPath(docPath)

	printBrandHeader()
	printRow(10, "preview", underline(url))
	printRow(10, "document", display)
	fmt.Println()
	fmt.Printf("  %s\n\n", dim("press ctrl+c to stop"))
}
