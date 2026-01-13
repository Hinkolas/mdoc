package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hinkolas/mdoc/src"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(printCmd)

	// Print Command Flags
	printCmd.Flags().StringP("config", "c", "", "Path to config file")
	printCmd.Flags().StringP("output", "o", "", "Path of the output file")
	printCmd.Flags().Bool("html", false, "Also export the raw HTML file alongside the PDF")
}

var printCmd = &cobra.Command{
	Use:   "print [file]",
	Short: "Generates a PDF from the provided markdown document.",
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

		// Determine output file path
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		outputPath := strings.TrimSuffix(base, ext) + ".pdf"

		// Check if HTML export is requested
		exportHTML, _ := cmd.Flags().GetBool("html")

		// Save document to output file
		err = document.Save(outputPath, exportHTML)
		if err != nil {
			fmt.Println("Error rendering document:", err)
			os.Exit(1)
		}

		os.Exit(0)

	},
}
