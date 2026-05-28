package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hinkolas/mdoc/internal/browser"
)

var installRevision int

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download a compatible Chromium snapshot into the user cache directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		bin, err := browser.Install(installRevision)
		if err != nil {
			return err
		}
		fmt.Printf("chromium installed: %s\n", bin)
		return nil
	},
}

func init() {
	installCmd.Flags().IntVar(&installRevision, "chromium", -1, "Specific Chromium revision to install (default: latest known)")
	rootCmd.AddCommand(installCmd)
}
