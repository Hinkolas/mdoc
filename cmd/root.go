package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is the build version. It is "dev" for local builds and is overridden
// at release time by GoReleaser via -ldflags -X.
var Version = "dev"

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
