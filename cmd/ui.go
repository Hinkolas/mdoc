package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/hinkolas/mdoc/internal/paths"
)

// stdoutIsTTY reports whether stdout is attached to a terminal. Set once
// at init so all the styling helpers can cheaply skip ANSI escapes when
// the user is piping output to a file.
var stdoutIsTTY = isTTY(os.Stdout)

// stdinIsTTY reports whether stdin is attached to a terminal — i.e. whether
// there's a human we can ask an interactive question of.
var stdinIsTTY = isTTY(os.Stdin)

func isTTY(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// confirmOverwrite decides whether a command may write to outPath when it
// might already exist. With force it always may. Otherwise, if the file is
// already there, it asks on an interactive terminal and refuses outright
// when there's no TTY to ask on — so a script never silently clobbers an
// existing artifact. Returns whether to proceed.
func confirmOverwrite(outPath string, force bool) (bool, error) {
	if force {
		return true, nil
	}
	if _, err := os.Stat(outPath); err != nil {
		// Doesn't exist (or can't be stat'd) — nothing to confirm; let the
		// write itself surface any real error.
		return true, nil
	}
	if !stdinIsTTY {
		return false, fmt.Errorf("%s already exists; pass --force to overwrite", displayPath(outPath))
	}
	fmt.Fprintf(os.Stderr, "%s %s already exists. Overwrite? %s ",
		yellow("?"), bold(displayPath(outPath)), dim("[y/N]"))
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// printCancelled notes that the user declined an overwrite and the existing
// file was left untouched.
func printCancelled(outPath string) {
	fmt.Fprintf(os.Stderr, "%s cancelled — %s left unchanged\n", red("✗"), displayPath(outPath))
}

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
	printRowMarked(cyan("➜"), labelWidth, label, value)
}

// printRowMarked is printRow with a custom leading marker — e.g. a yellow ⚠
// for a warning row — so a diagnostic can sit inside the banner block and
// still line up with the arrow rows.
func printRowMarked(marker string, labelWidth int, label, value string) {
	pad := strings.Repeat(" ", labelWidth-len(label))
	fmt.Printf("  %s  %s%s%s\n", marker, dim(label), pad, value)
}

// displayPath formats a path for the banners: absolute, with the home
// directory collapsed to "~". Consistent everywhere — see paths.Display.
func displayPath(p string) string { return paths.Display(p) }

// The live-log helpers below print Vite-style timestamped event lines while a
// preview session is running, e.g.:
//
//	10:24:14  reloaded  ~/Github/mdoc/example/document.md
//	10:24:14  warning   theme "test" not found in …; using the built-in "system" theme
//
// Lines are flush-left (the indented banner is the header; the log is the
// stream) with the label column padded so details align. They go to stderr so
// stdout stays clean for any piped path output.
func logTime() string { return time.Now().Format("15:04:05") }

func logEvent(label string, color func(string) string, detail string) {
	fmt.Fprintf(os.Stderr, "%s  %s  %s\n", dim(logTime()), color(fmt.Sprintf("%-8s", label)), detail)
}

func logReload(detail string) { logEvent("reloaded", cyan, dim(detail)) }
func logReady(detail string)  { logEvent("ready", green, dim(detail)) }
func logLiveWarn(msg string)  { logEvent("warning", yellow, msg) }
func logLiveErr(msg string)   { logEvent("error", red, msg) }

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
