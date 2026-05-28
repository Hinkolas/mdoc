package cmd

import (
	"fmt"
	"os"
	"time"

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
		start := time.Now()
		out, err := print.Print(doc, thm, print.Options{
			OutputPath: printOutput,
			WriteHTML:  printHTMLOut,
			Version:    Version,
		})
		if err != nil {
			return err
		}
		dur := time.Since(start)

		// In a pipe, just emit the absolute path on stdout so scripts can
		// chain commands like `mdoc print foo.md | xargs open`. In a TTY
		// the path doesn't go to stdout — only the banner does.
		if !stdoutIsTTY {
			fmt.Println(out)
			return nil
		}
		printPrintBanner(doc.Path, out, dur)
		return nil
	},
}

func init() {
	printCmd.Flags().StringVarP(&printOutput, "output", "o", "", "Output PDF path (default: <input>.pdf)")
	printCmd.Flags().BoolVar(&printHTMLOut, "html", false, "Also write the rendered HTML alongside the PDF")
	rootCmd.AddCommand(printCmd)
}

func printPrintBanner(srcPath, outPath string, dur time.Duration) {
	src := relToCwd(srcPath)
	dst := relToCwd(outPath)

	size := ""
	if fi, err := os.Stat(outPath); err == nil {
		size = humanSize(fi.Size())
	}

	meta := ""
	if size != "" {
		meta = fmt.Sprintf("  %s", dim(fmt.Sprintf("(%s · %s)", size, shortDuration(dur))))
	} else {
		meta = fmt.Sprintf("  %s", dim(fmt.Sprintf("(%s)", shortDuration(dur))))
	}

	printBrandHeader()
	printRow(8, "source", src)
	printRow(8, "output", dst+meta)
	fmt.Println()
}

