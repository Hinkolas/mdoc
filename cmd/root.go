package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:           "mdoc",
	Short:         "Render markdown documents to PDF with paged.js-driven pagination.",
	Version:       fmt.Sprintf("%s, %s/%s", Version, runtime.GOOS, runtime.GOARCH),
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
