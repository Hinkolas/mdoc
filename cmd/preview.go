package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hinkolas/mdoc/src/core"
	"github.com/hinkolas/mdoc/src/preview"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(previewCmd)

	previewCmd.Flags().IntP("port", "p", 3141, "Port to run the preview server on")
}

var previewCmd = &cobra.Command{
	Use:   "preview [file]",
	Short: "Starts a local web server to preview the markdown document.",
	Run: func(cmd *cobra.Command, args []string) {

		// Determine input file path
		var inputPath string
		if len(args) > 0 {
			inputPath = args[0]
		} else {
			fmt.Println("No input file provided")
			os.Exit(1)
		}

		port, _ := cmd.Flags().GetInt("port")

		// Parse the document initially to get theme path for watching
		document, err := core.OpenDocument(inputPath)
		if err != nil {
			fmt.Printf("Error opening document: %v\n", err)
			os.Exit(1)
		}

		// Create the preview server
		server := preview.NewPreviewServer(inputPath)

		// Create file watcher for source and theme files
		watcher, err := preview.NewWatcher(inputPath, document.ThemePath)
		if err != nil {
			fmt.Printf("Warning: Failed to create file watcher: %v\n", err)
			// Continue without file watching
		} else {
			// Set up change notification
			watcher.OnChanged = func() {
				if err := server.SendEvent("source_changed"); err != nil {
					// Client might not be connected yet, that's okay
					fmt.Printf("Could not send source_changed event: %v\n", err)
				}
			}

			// Start watching in background
			go watcher.Watch()
			defer watcher.Close()
		}

		// Handle shutdown signals
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sig
			fmt.Println("\nShutting down...")
			os.Exit(0)
		}()

		// Start the server (blocks)
		fmt.Printf("Preview server starting on http://localhost:%d/preview\n", port)
		if err := server.Start(port); err != nil {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}

	},
}
