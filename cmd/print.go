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
	printForce   bool
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
		outPath, err := print.ResolveOutputPath(doc, printOutput)
		if err != nil {
			return err
		}
		proceed, err := confirmOverwrite(outPath, printForce)
		if err != nil {
			return err
		}
		if !proceed {
			printCancelled(outPath)
			return nil
		}

		thm, twarn := theme.Resolve(doc.Config.Theme, doc.Dir)
		start := time.Now()
		out, err := print.Print(doc, thm, print.Options{
			OutputPath: outPath,
			WriteHTML:  printHTMLOut,
			Version:    Version,
		})
		if err != nil {
			return err
		}
		dur := time.Since(start)

		// In a pipe, just emit the absolute path on stdout so scripts can
		// chain commands like `mdoc print foo.md | xargs open`. In a TTY
		// the path doesn't go to stdout — only the banner does. A theme
		// fallback is folded into the banner in a TTY; in a pipe it goes to
		// stderr (full detail) so it neither pollutes stdout nor is lost.
		if !stdoutIsTTY {
			fmt.Println(out)
			if twarn != nil {
				printWarn(twarn.Error())
			}
			return nil
		}
		printPrintBanner(doc.Path, out, dur, twarn)
		return nil
	},
}

func init() {
	printCmd.Flags().StringVarP(&printOutput, "output", "o", "", "Output PDF path (default: <input>.pdf)")
	printCmd.Flags().BoolVar(&printHTMLOut, "html", false, "Also write the rendered HTML alongside the PDF")
	printCmd.Flags().BoolVarP(&printForce, "force", "f", false, "Overwrite the output file if it already exists")
	rootCmd.AddCommand(printCmd)
}

func printPrintBanner(srcPath, outPath string, dur time.Duration, themeWarn error) {
	src := displayPath(srcPath)
	dst := displayPath(outPath)

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
	// On a theme fallback, slot a concise warning row into the banner.
	if fb, ok := themeWarn.(*theme.Fallback); ok {
		printRowMarked(yellow("⚠"), 8, "theme", fb.Short())
	} else if themeWarn != nil {
		printRowMarked(yellow("⚠"), 8, "theme", themeWarn.Error())
	}
	fmt.Println()
}

