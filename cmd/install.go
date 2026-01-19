package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-rod/rod/lib/launcher"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)

	// debug Command Flags
	installCmd.Flags().StringP("config", "c", "", "Path to config file")
	installCmd.Flags().String("tailwind", "", "Version of Tailwind CSS to install")
	installCmd.Flags().Int("chromium", -1, "Revision of Chromium to install")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs latest snapshot of Chromium",
	Run: func(cmd *cobra.Command, args []string) {

		revision, _ := cmd.Flags().GetInt("chromium")
		err := installBrowser(revision)
		if err != nil {
			fmt.Println("Failed to install browser:", err)
			os.Exit(1)
		}

		tailwind, _ := cmd.Flags().GetString("tailwind")
		err = installTailwind(tailwind)
		if err != nil {
			fmt.Println("Failed to install tailwind:", err)
			os.Exit(1)
		}

		fmt.Println("Browser downloaded successfully. You are ready to go!")

	},
}

func installBrowser(revision int) error {

	// Look for the system browser path ('chromium', 'google-chrome', etc.)
	path, has := launcher.LookPath()
	if has {
		return fmt.Errorf("browser already installed at: %s", path)
	}

	userConfig, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %w", err)
	}

	// Create the browser launcher instance
	browser := launcher.NewBrowser()
	browser.RootDir = filepath.Join(userConfig, "mdoc", "chromium")

	// Check if revision version is set
	if revision != -1 {
		browser.Revision = revision
	}

	fmt.Println("Downloading Browser...")
	err = browser.Download()
	if err != nil {
		return fmt.Errorf("failed to download browser: %w", err)
	}

	return nil

}

func installTailwind(version string) error {

	// Default to version 3.4.17 if not specified
	if version == "" {
		version = "3.4.17"
	}

	// Download tailwind
	downloadURL := fmt.Sprintf("https://cdn.tailwindcss.com/%s", version)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to fetch tailwind: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download tailwind: HTTP %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read tailwind response: %w", err)
	}

	// Get user config directory and create tailwind directory
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %w", err)
	}

	tailwindDir := filepath.Join(userConfig, "mdoc", "tailwindcss")
	err = os.MkdirAll(tailwindDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create tailwind directory: %w", err)
	}

	// Write the tailwind file as [version]
	tailwindFile := filepath.Join(tailwindDir, version)
	err = os.WriteFile(tailwindFile, body, 0644)
	if err != nil {
		return fmt.Errorf("failed to write tailwind file: %w", err)
	}

	fmt.Printf("Tailwind CSS %s installed to: %s\n", version, tailwindFile)

	return nil
}
