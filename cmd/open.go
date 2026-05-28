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
		fmt.Printf("preview: %s\n", srv.URL())

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
	openCmd.Flags().IntVarP(&openPort, "port", "p", 3141, "Preview server port (0 = pick a free port)")
	rootCmd.AddCommand(openCmd)
}
