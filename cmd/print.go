package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/print"
	"github.com/hinkolas/mdoc/internal/theme"
)

var (
	printOutput  string
	printHTMLOut bool
)

var printCmd = &cobra.Command{
	Use:   "print <file>",
	Short: "Render a markdown document to PDF.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := document.Open(args[0])
		if err != nil {
			return err
		}
		thm, err := theme.Resolve(doc.Config.Theme, doc.Dir)
		if err != nil {
			return err
		}
		out, err := print.Print(doc, thm, print.Options{
			OutputPath: printOutput,
			WriteHTML:  printHTMLOut,
			Version:    Version,
		})
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	printCmd.Flags().StringVarP(&printOutput, "output", "o", "", "Output PDF path (default: <input>.pdf)")
	printCmd.Flags().BoolVar(&printHTMLOut, "html", false, "Also write the rendered HTML alongside the PDF")
	rootCmd.AddCommand(printCmd)
}
