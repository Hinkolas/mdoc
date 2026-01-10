package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
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
	rootCmd.AddCommand(startCmd)

	// Start Command Flags
	startCmd.Flags().StringP("config", "c", "config.yaml", "Path to config file")
}

var startCmd = &cobra.Command{
	Use:   "debug",
	Short: "This creates a pdf of a debug page with some system information to troubleshoot errors or validate the installation.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Initializing Browser...")

		// 1. Look for the system browser path ('chromium', 'google-chrome', etc.)
		path, has := launcher.LookPath()
		if !has {
			fmt.Println("Browser not found!")
			os.Exit(1)
		}

		// 2. This attempts to launch the found Chrome installation.
		u, err := launcher.New().Bin(path).Launch()
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
		outputPath := fmt.Sprintf("debug-%d.pdf", time.Now().Unix())
		_ = utils.OutputFile(outputPath, pdfData)

		fmt.Printf("Done! Saved to %s\n", outputPath)

		os.Exit(0)

	},
}
