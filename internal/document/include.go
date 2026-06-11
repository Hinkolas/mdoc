package document

// Include resolution splices `:::include <path>` directives into the document
// body at the source-text level, before goldmark ever parses it. This is the
// LaTeX `\input` model: a 100-page document can be split into one markdown file
// per chapter and stitched back together from a root/index file.
//
// Why textual splicing rather than rendering each file and concatenating the
// HTML: the whole apparatus — heading numbering, the TOC, `[#id]`
// cross-references, per-chapter figure/table numbering, `[@key]` citations and
// the bibliography — is computed in a single transform pass over the combined
// AST (see internal/mdext). Continuous numbering across chapters, a reference in
// chapter 5 pointing at a figure in chapter 2, and a document-wide TOC only work
// if the parser sees one combined source. So includes are flattened here, up
// front, and the rest of the pipeline is untouched.
//
// Included files may carry their own YAML frontmatter so they stay individually
// openable with `mdoc open chapter.md`; on include that frontmatter is parsed
// off and discarded — all configuration comes from the root document. Relative
// asset paths inside an included file resolve relative to the ROOT document's
// directory (the combined body is rendered as if it all lived there), so shared
// images belong under the root's tree.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adrg/frontmatter"
)

// maxIncludeDepth bounds nested includes so a deep (or pathological) tree can't
// blow the stack; cycles are caught separately and exactly.
const maxIncludeDepth = 64

// resolveIncludes returns body with every `:::include <path>` line replaced by
// the (recursively resolved) body of the referenced file, plus the absolute
// paths of all files pulled in, in include order. baseDir is the directory
// include paths in this body resolve against (the directory of the file the
// body came from). stack is the chain of absolute paths currently being
// resolved, starting with the root document, used for cycle detection.
func resolveIncludes(body, baseDir string, stack []string) (string, []string, error) {
	if len(stack) > maxIncludeDepth {
		return "", nil, fmt.Errorf("include depth exceeds %d (cycle or runaway nesting near %s)", maxIncludeDepth, baseDir)
	}

	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	var included []string

	var fenceChar byte // 0 when not inside a fenced code block
	var fenceLen int

	for _, raw := range lines {
		line := strings.TrimSuffix(raw, "\r")
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)

		// Code fences are passed through verbatim so a documented `:::include`
		// inside a ``` block isn't treated as a real include.
		if fenceChar != 0 {
			if indent <= 3 && closesFence(trimmed, fenceChar, fenceLen) {
				fenceChar = 0
			}
			out = append(out, line)
			continue
		}
		if indent <= 3 {
			if c, n := opensFence(trimmed); n > 0 {
				fenceChar, fenceLen = c, n
				out = append(out, line)
				continue
			}
		}

		path, ok := includeTarget(trimmed, indent)
		if !ok {
			out = append(out, line)
			continue
		}

		abs := path
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(baseDir, abs)
		}
		abs, err := filepath.Abs(abs)
		if err != nil {
			return "", nil, fmt.Errorf("resolve include %q: %w", path, err)
		}

		if slices.Contains(stack, abs) {
			return "", nil, fmt.Errorf("include cycle: %s includes itself (via %s)", stack[0], abs)
		}

		childBody, err := readIncludedBody(abs)
		if err != nil {
			return "", nil, fmt.Errorf("include %q (from %s): %w", path, stack[len(stack)-1], err)
		}
		childCombined, childIncluded, err := resolveIncludes(childBody, filepath.Dir(abs), append(stack, abs))
		if err != nil {
			return "", nil, err
		}

		// Surround the spliced content with blank lines so adjacent markdown
		// blocks (a trailing paragraph here, a leading heading there) don't
		// accidentally merge across the seam.
		out = append(out, "")
		out = append(out, strings.Split(childCombined, "\n")...)
		out = append(out, "")

		included = append(included, abs)
		included = append(included, childIncluded...)
	}

	return strings.Join(out, "\n"), included, nil
}

// includeTarget reports whether a code-fence-cleared, leading-space-trimmed line
// is a `:::include <path>` directive and, if so, returns the path. indent is the
// number of leading spaces that were stripped; more than three would make the
// line an indented code block rather than a directive. The path is everything
// after the directive name (trimmed), so paths with spaces work without quoting.
func includeTarget(trimmed string, indent int) (string, bool) {
	if indent > 3 {
		return "", false
	}
	i := 0
	for i < len(trimmed) && trimmed[i] == ':' {
		i++
	}
	if i < 3 {
		return "", false
	}
	rest := strings.TrimSpace(trimmed[i:])
	const name = "include"
	if rest == name {
		return "", false // `:::include` with no path — treated as not-an-include here
	}
	after, ok := strings.CutPrefix(rest, name)
	if !ok || (after[0] != ' ' && after[0] != '\t') {
		return "", false // e.g. `:::includes` or `:::toc`
	}
	return strings.TrimSpace(after), true
}

// readIncludedBody reads an included file and strips any YAML frontmatter,
// returning just the markdown body. The frontmatter is discarded: configuration
// always comes from the root document.
func readIncludedBody(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var ignored struct{} // discard the included file's frontmatter
	body, err := frontmatter.Parse(f, &ignored)
	if err != nil {
		return "", fmt.Errorf("parse frontmatter: %w", err)
	}
	return string(body), nil
}

// opensFence reports the fence character and run length if trimmed opens a
// fenced code block (three or more backticks or tildes), else (0, 0). A backtick
// fence may not carry a backtick in its info string (CommonMark).
func opensFence(trimmed string) (byte, int) {
	if trimmed == "" {
		return 0, 0
	}
	c := trimmed[0]
	if c != '`' && c != '~' {
		return 0, 0
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == c {
		n++
	}
	if n < 3 {
		return 0, 0
	}
	if c == '`' && strings.ContainsRune(trimmed[n:], '`') {
		return 0, 0
	}
	return c, n
}

// closesFence reports whether trimmed is a closing fence for an open block of
// the given character and length: at least minLen of that character and nothing
// else but trailing whitespace.
func closesFence(trimmed string, char byte, minLen int) bool {
	n := 0
	for n < len(trimmed) && trimmed[n] == char {
		n++
	}
	if n < minLen {
		return false
	}
	return strings.TrimSpace(trimmed[n:]) == ""
}
