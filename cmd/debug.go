package cmd

// import (
// 	"bytes"
// 	_ "embed"
// 	"fmt"
// 	"html/template"
// 	"os"

// 	"github.com/hinkolas/mdoc/src/core"

// 	"github.com/spf13/cobra"
// )

// //go:embed debug.html
// var debugTemplateDef string

// func init() {
// 	rootCmd.AddCommand(debugCmd)

// 	// debug Command Flags
// 	debugCmd.Flags().StringP("config", "c", "", "Path to config file")
// 	debugCmd.Flags().StringP("output", "o", "", "Path of the output file")
// }

// // TODO: Adjust command and template to use the document.Save() function for DRY
// var debugCmd = &cobra.Command{
// 	Use:   "debug",
// 	Short: "This creates a pdf of a debug page with some system information to troubleshoot errors or validate the installation.",
// 	Run: func(cmd *cobra.Command, args []string) {

// 		document, err := core.NewDocument()
// 		if err != nil {
// 			fmt.Println("Error creating document:", err)
// 			os.Exit(1)
// 		}

// 		data, err := document.RenderData()
// 		if err != nil {
// 			fmt.Println("Error collecting render data:", err)
// 			os.Exit(1)
// 		}

// 		debugTemplate, err := template.New("debug").Parse(debugTemplateDef)
// 		if err != nil {
// 			fmt.Println("Error parsing debug template:", err)
// 			os.Exit(1)
// 		}

// 		var htmlBuf bytes.Buffer
// 		if err := debugTemplate.Execute(&htmlBuf, data); err != nil {
// 			fmt.Println("Error executing debug template:", err)
// 			os.Exit(1)
// 		}

// 		err = document.Print("./debug.pdf")
// 		if err != nil {
// 			fmt.Println("Error printing document:", err)
// 			os.Exit(1)
// 		}

// 		os.Exit(0)

// 	},
// }
