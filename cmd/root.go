package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/hinkolas/mdoc/src"
	"github.com/spf13/cobra"
)

func init() {
	// Start Command Flags
	rootCmd.Flags().StringP("config", "c", "config.yaml", "Path	 to config file (.yml)")
	rootCmd.Flags().StringP("output", "o", "", "Path of the output file")
	rootCmd.Flags().StringP("template", "t", "", "Path of the template file (.html)")
}

var rootCmd = &cobra.Command{
	Version: fmt.Sprintf("%s, %s/%s", "0.0.1", runtime.GOOS, runtime.GOARCH),
	Use:     "mdoc",
	Short:   "An easy to use cli tool for turning your markdown files into good looking PDFs with customizable templates and more. ",
	Run: func(cmd *cobra.Command, args []string) {

		// Return error if no file was provided
		if len(args) != 1 {
			fmt.Println("Please provide exactly 1 markdown file to convert.")
			os.Exit(1)
		}

		inputPath := args[0]
		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		input, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		src.Create(string(input))

		os.Exit(0)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

}
