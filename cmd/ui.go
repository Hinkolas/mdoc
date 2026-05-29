package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
func yellow(s string) string    { return ansi("33", s) }
func underline(s string) string { return ansi("4;36", s) }

// printWarn writes a non-fatal warning to stderr. Used for things the user
// should know about but that don't stop the command — e.g. a named theme
// that couldn't be found, where rendering falls back to the built-in default.
func printWarn(msg string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", yellow("warning:"), msg)
}

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

// relToCwd returns a path made relative to the current working directory
// when that's shorter and stays inside the cwd subtree; otherwise the
// absolute path is returned unchanged.
func relToCwd(p string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return p
	}
	rel, err := filepath.Rel(cwd, p)
	if err != nil || strings.HasPrefix(rel, "..") {
		return p
	}
	return rel
}

// humanSize formats a byte count as "267 KB", "1.4 MB", etc. — the kind
// of unit a user actually cares about for a generated artifact.
func humanSize(n int64) string {
	const k = 1024.0
	f := float64(n)
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.0f KB", f/k)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", f/(k*k))
	default:
		return fmt.Sprintf("%.1f GB", f/(k*k*k))
	}
}

// shortDuration trims the noise from time.Duration's String — ms for very
// fast operations, one-decimal seconds otherwise.
func shortDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
