package cmd

import (
	"fmt"
	"os"
	"strings"
)

// stdoutIsTTY reports whether stdout is attached to a terminal. Set once
// at init so all the styling helpers can cheaply skip ANSI escapes when
// the user is piping output to a file.
var stdoutIsTTY = func() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}()

func ansi(code, text string) string {
	if !stdoutIsTTY {
		return text
	}
	return "\033[" + code + "m" + text + "\033[0m"
}

func bold(s string) string      { return ansi("1", s) }
func dim(s string) string       { return ansi("2", s) }
func cyan(s string) string      { return ansi("36", s) }
func green(s string) string     { return ansi("32", s) }
func red(s string) string       { return ansi("31", s) }
func underline(s string) string { return ansi("4;36", s) }

// printBrandHeader prints the "  mdoc  v0.1.0" line surrounded by blank
// lines that every command shares as its header.
func printBrandHeader() {
	fmt.Println()
	fmt.Printf("  %s  %s\n", bold("mdoc"), dim("v"+Version))
	fmt.Println()
}

// printRow prints "  ➜  label<pad>value" with the arrow accent and a
// dimmed label. labelWidth pads the label column so multiple rows line up.
func printRow(labelWidth int, label, value string) {
	pad := strings.Repeat(" ", labelWidth-len(label))
	fmt.Printf("  %s  %s%s%s\n", cyan("➜"), dim(label), pad, value)
}
