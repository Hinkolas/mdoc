package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/bundle"
	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/theme"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Bundle a document, its theme, and its assets into a portable .mdoc zip.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := document.Open(args[0])
		if err != nil {
			return err
		}
		thm, twarn := theme.Resolve(doc.Config.Theme, doc.Dir)

		start := time.Now()
		res, err := bundle.Export(doc, thm, bundle.Options{OutputPath: exportOutput})
		if err != nil {
			return err
		}
		dur := time.Since(start)

		// Keep stdout machine-readable in a pipe (so `mdoc export foo.md
		// | xargs <thing>` still works); only show the banner in a tty. A
		// theme fallback is folded into the banner in a TTY; in a pipe it
		// goes to stderr (full detail) so it neither pollutes stdout nor is lost.
		if !stdoutIsTTY {
			fmt.Println(res.OutputPath)
			if twarn != nil {
				printWarn(twarn.Error())
			}
			return nil
		}
		printExportBanner(doc.Path, res, dur, twarn)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output bundle path (default: <input>.mdoc)")
	rootCmd.AddCommand(exportCmd)
}

func printExportBanner(srcPath string, res *bundle.Result, dur time.Duration, themeWarn error) {
	src := displayPath(srcPath)
	dst := displayPath(res.OutputPath)

	size := ""
	if fi, err := os.Stat(res.OutputPath); err == nil {
		size = humanSize(fi.Size())
	}

	parts := fmt.Sprintf("%d %s", len(res.Entries), plural(len(res.Entries), "file", "files"))
	meta := ""
	if size != "" {
		meta = "  " + dim(fmt.Sprintf("(%s · %s · %s)", size, parts, shortDuration(dur)))
	} else {
		meta = "  " + dim(fmt.Sprintf("(%s · %s)", parts, shortDuration(dur)))
	}

	printBrandHeader()
	printRow(8, "source", src)
	printRow(8, "bundle", dst+meta)
	// On a theme fallback, slot a concise warning row into the banner.
	if fb, ok := themeWarn.(*theme.Fallback); ok {
		printRowMarked(yellow("⚠"), 8, "theme", fb.Short())
	} else if themeWarn != nil {
		printRowMarked(yellow("⚠"), 8, "theme", themeWarn.Error())
	}
	fmt.Println()
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}
