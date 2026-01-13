package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-rod/rod/lib/launcher"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)

	// debug Command Flags
	installCmd.Flags().StringP("config", "c", "", "Path to config file")
	installCmd.Flags().IntP("revision", "r", -1, "Set a specific revision")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs latest snapshot of Chromium",
	Run: func(cmd *cobra.Command, args []string) {

		// Look for the system browser path ('chromium', 'google-chrome', etc.)
		path, has := launcher.LookPath()
		if has {
			fmt.Println("Browser already installed at: ", path)
			os.Exit(1)
		}

		userConfig, err := os.UserConfigDir()
		if err != nil {
			fmt.Println("Failed to get user config directory:", err)
			os.Exit(1)
		}

		// Create the browser launcher instance
		browser := launcher.NewBrowser()
		browser.RootDir = filepath.Join(userConfig, "mdoc", "chromium")

		// Check if revision version is set
		if revision, _ := cmd.Flags().GetInt("revision"); revision != -1 {
			fmt.Println("Downloading snapshot ", revision)
			browser.Revision = revision
		}

		fmt.Println("Downloading Browser...")
		err = browser.Download()
		if err != nil {
			fmt.Println("Failed to download browser:", err)
			os.Exit(1)
		}

		fmt.Println("Browser downloaded successfully. You are ready to go!")

	},
}
