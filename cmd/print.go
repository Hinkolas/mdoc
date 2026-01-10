package cmd

import (
	_ "embed"
	"fmt"
	"os"

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

		fmt.Println("TODO: Generate PDF")
		os.Exit(0)

	},
}
