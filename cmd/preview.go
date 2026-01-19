package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/hinkolas/mdoc/src"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(previewCmd)

	// debug Command Flags
	previewCmd.Flags().StringP("config", "c", "", "Path to config file")
	previewCmd.Flags().StringP("output", "o", "", "Path of the output file")
}

var previewCmd = &cobra.Command{
	Use:   "preview [file]",
	Short: "Starts a local web server to preview the markdown document.",
	Run: func(cmd *cobra.Command, args []string) {

		var inputPath string

		// Determine input file path
		if len(args) > 0 {
			inputPath = args[0]
		} else {
			fmt.Println("No input file provided")
			os.Exit(1)
		}

		// Open file
		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			os.Exit(1)
		}
		defer file.Close()

		// Parse markdown
		document, err := src.ParseDocument(file)
		if err != nil {
			fmt.Println("Error parsing document:", err)
			os.Exit(1)
		}

		userConfig, err := os.UserConfigDir()
		if err != nil {
			fmt.Println("Failed to get user config directory:", err)
			os.Exit(1)
		}

		// Create the browser launcher instance
		chromium := launcher.NewBrowser()
		chromium.RootDir = filepath.Join(userConfig, "mdoc", "chromium")

		var binPath string
		err = chromium.Validate()
		if err != nil {
			fmt.Println("Unable to find packaged browser. Looking for local alternative...")
			var has bool = false
			binPath, has = launcher.LookPath()
			if !has {
				fmt.Println("No compatible browser found! Please run `mdoc install` to download the latest chromium snapshot.")
				os.Exit(1)
			}
		} else {
			binPath = chromium.BinPath()
		}

		fmt.Println("Initializing Browser...")

		// 2. This attempts to launch the found Chrome installation.
		u, err := launcher.New().Bin(binPath).Headless(false).Launch()
		if err != nil {
			fmt.Println("Failed to launch browser:", err)
			os.Exit(1)
		}

		// 3. Connect to the browser
		browser := rod.New().ControlURL(u).MustConnect()
		defer browser.MustClose()

		// 4. Create a page (tab)
		page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create page: %v\n", err)
			os.Exit(1)
		}

		// // TODO: Remember window position and size and restore them when the browser is closed
		// left := 100
		// top := 100
		// width := 1200
		// height := 800
		// err = page.SetWindow(&proto.BrowserBounds{
		// 	Left:   &left,
		// 	Top:    &top,
		// 	Width:  &width,
		// 	Height: &height,
		// })

		// Initial render
		body, err := document.Render(src.RenderModePreview)
		if err != nil {
			fmt.Println("Failed to render document:", err)
			os.Exit(1)
		}
		page.MustSetDocumentContent(body)

		// Set up file watcher for hot reload
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Println("Failed to create file watcher:", err)
			os.Exit(1)
		}
		defer watcher.Close()

		// Watch the source file
		absInputPath, err := filepath.Abs(inputPath)
		if err != nil {
			fmt.Println("Failed to get absolute path for input file:", err)
			os.Exit(1)
		}
		if err := watcher.Add(absInputPath); err != nil {
			fmt.Println("Failed to watch input file:", err)
			os.Exit(1)
		}

		// Watch the theme file
		themePath := filepath.Join(os.ExpandEnv(src.THEME_DIR), document.Config.Theme+".html")
		absThemePath, err := filepath.Abs(themePath)
		if err != nil {
			fmt.Println("Failed to get absolute path for theme file:", err)
			os.Exit(1)
		}
		if err := watcher.Add(absThemePath); err != nil {
			fmt.Printf("Warning: Failed to watch theme file %s: %v\n", absThemePath, err)
			// Continue anyway - theme watching is optional
		}

		fmt.Printf("Watching files for changes:\n  - %s\n  - %s\n", absInputPath, absThemePath)
		fmt.Println("Preview is running. Close the browser window or press Ctrl+C to exit.")

		// Wait for either browser close or interrupt signal
		done := make(chan struct{})
		go func() {
			browser.WaitEvent(&proto.TargetTargetDestroyed{})()
			close(done)
		}()

		// TODO: Rework/clean up this code
		// File watcher goroutine
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					// Only react to write events
					if event.Op&fsnotify.Write == fsnotify.Write {
						fmt.Printf("File changed: %s. Reloading...\n", event.Name)

						// Re-parse the document
						file, err := os.Open(inputPath)
						if err != nil {
							fmt.Println("Error opening file:", err)
							continue
						}

						newDoc, err := src.ParseDocument(file)
						file.Close()
						if err != nil {
							fmt.Println("Error parsing document:", err)
							continue
						}

						// Re-render and update the page
						body, err := newDoc.Render(src.RenderModePreview)
						if err != nil {
							fmt.Println("Failed to render document:", err)
							continue
						}

						page.MustSetDocumentContent(body)
						fmt.Println("Reload complete.")
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					fmt.Println("Watcher error:", err)
				case <-done:
					return
				}
			}
		}()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		select {
		case <-done:
			// Browser was closed
		case <-sig:
			// User pressed Ctrl+C
		}

	},
}
