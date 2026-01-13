package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/goforj/godump"
	"github.com/hinkolas/mdoc/src"

	"github.com/spf13/cobra"
)

//go:embed debug.html
var tmplDef string

func init() {
	rootCmd.AddCommand(debugCmd)

	// debug Command Flags
	debugCmd.Flags().StringP("config", "c", "", "Path to config file")
	debugCmd.Flags().StringP("output", "o", "", "Path of the output file")
}

// TODO: Adjust command and template to use the document.Save() function for DRY
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "This creates a pdf of a debug page with some system information to troubleshoot errors or validate the installation.",
	Run: func(cmd *cobra.Command, args []string) {

		userConfig, err := os.UserConfigDir()
		if err != nil {
			fmt.Println("Failed to get user config directory:", err)
			os.Exit(1)
		}

		// Create the browser launcher instance
		chromium := launcher.NewBrowser()
		chromium.RootDir = filepath.Join(userConfig, "mdoc", "chromium")

		var binPath string
		err = chromium.Validate()
		if err != nil {
			fmt.Println("Unable to find packaged browser. Looking for local alternative...")
			var has bool = false
			binPath, has = launcher.LookPath()
			if !has {
				fmt.Println("No compatible browser found! Please run `mdoc install` to download the latest chromium snapshot.")
				os.Exit(1)
			}
		} else {
			binPath = chromium.BinPath()
		}

		fmt.Println("Initializing Browser...")

		// 2. This attempts to launch the found Chrome installation.
		u, err := launcher.New().Bin(binPath).Launch()
		if err != nil {
			fmt.Println("Failed to launch browser:", err)
			os.Exit(1)
		}

		// 3. Connect to the browser
		browser := rod.New().ControlURL(u).MustConnect()
		defer browser.MustClose()

		// 4. Create a page (tab)
		page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create page: %v\n", err)
			os.Exit(1)
		}

		// 5. Render the go html template with nil object for now
		tmpl, err := template.New("debug").Parse(tmplDef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %v\n", err)
			os.Exit(1)
		}

		// 6. Collect system data for providing it to the template execution
		systemData, err := src.CollectSystemData()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to collect system data: %v\n", err)
			os.Exit(1)
		}

		// TODO: Turn this int a static type later on
		tmplData := map[string]any{
			"System": systemData,
			"Body":   godump.DumpJSONStr(systemData),
		}

		// 7. Execute template into memory buffer
		var buf bytes.Buffer

		// We pass 'nil' for data as requested, but this is where your struct would go.
		if err := tmpl.Execute(&buf, tmplData); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute template: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Loading HTML into browser...")

		// Load the content into the page
		// Rod expects a string, so we convert the buffer contents.
		page.MustSetDocumentContent(buf.String())
		page.MustWaitStable()

		fmt.Println("Generating PDF...")

		// 8. PDF Options
		// You can strictly type the config using the proto package
		paperWidth := 8.27   // A4 Width (inches)
		paperHeight := 11.69 // A4 Height (inches)
		margin := 0.5
		printBg := true
		pdfData, err := page.PDF(&proto.PagePrintToPDF{
			PaperWidth:      &paperWidth,
			PaperHeight:     &paperHeight,
			MarginTop:       &margin,
			MarginBottom:    &margin,
			MarginLeft:      &margin,
			MarginRight:     &margin,
			PrintBackground: printBg, // Important for CSS backgrounds!
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate PDF: %v\n", err)
			os.Exit(1)
		}

		// 9. Save to disk
		outputPath := fmt.Sprintf(".local/debug-%d.pdf", time.Now().Unix())
		_ = utils.OutputFile(outputPath, pdfData)

		fmt.Printf("Done! Saved to %s\n", outputPath)

		os.Exit(0)

	},
}
