package cmd

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/hinkolas/mdoc/src"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(printCmd)

	// Print Command Flags
	printCmd.Flags().StringP("config", "c", "", "Path to config file")
	printCmd.Flags().StringP("output", "o", "", "Path of the output file")
}

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Generates a PDF from the provided markdown document.",
	Run: func(cmd *cobra.Command, args []string) {

		// Open file
		file, err := os.Open(".local/hello-world.md")
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

		err = document.Save("./test.pdf")
		if err != nil {
			fmt.Println("Error rendering document:", err)
			os.Exit(1)
		}

		os.Exit(0)

	},
}
