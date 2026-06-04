package mdext

import (
	"slices"
	"strconv"
	"strings"

	"github.com/hinkolas/mdoc/internal/document"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const defaultTOCDepth = 3

// transformer is the foundation pass: it numbers headings, collects them for
// any `:::toc`, numbers `[@key]` citations by first appearance, and attaches the
// resulting data onto the directive nodes (renderers get no parser.Context, so
// data must travel on the nodes).
type transformer struct {
	cfg Config
}

func newTransformer(cfg Config) *transformer { return &transformer{cfg: cfg} }

func (t *transformer) Transform(doc *gast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	refByKey := map[string]document.Reference{}
	for _, r := range t.cfg.References {
		if k := r.CiteKey(); k != "" {
			refByKey[k] = r
		}
	}

	// Pass 0: wrap `:::frontmatter` / `:::mainmatter` / `:::appendix` regions
	// into Matter containers (the numbering below reads the region per heading).
	wrapMatter(doc)

	// Pass 1: number headings, inject the section-number node, collect entries.
	// A heading's region sets the defaults (front = unnumbered + out of the TOC;
	// main = decimal; appendix = lettered); per-heading classes override them.
	var headings []HeadingEntry
	counters := make([]int, 7) // indices 1..6
	prevAppendix := false
	_ = gast.Walk(doc, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		h, ok := n.(*gast.Heading)
		if !ok || !entering {
			return gast.WalkContinue, nil
		}
		classes := classesOf(h)
		region := matterOf(h)

		numbered := t.cfg.Numbering.Enabled
		intoc := true
		if region == "front" {
			numbered, intoc = false, false
		}
		if slices.Contains(classes, "unnumbered") {
			numbered = false
		} else if slices.Contains(classes, "numbered") {
			numbered = t.cfg.Numbering.Enabled
		}
		if slices.Contains(classes, "notoc") {
			intoc = false
		} else if slices.Contains(classes, "intoc") {
			intoc = true
		}

		isAppendix := region == "appendix"
		if isAppendix && !prevAppendix {
			for i := range counters {
				counters[i] = 0
			}
		}
		prevAppendix = isAppendix

		title := nodeText(h, source) // before injecting the number
		number := ""
		if numbered && h.Level >= 1 && h.Level < len(counters) {
			counters[h.Level]++
			for i := h.Level + 1; i < len(counters); i++ {
				counters[i] = 0
			}
			number = formatNumber(counters, h.Level, isAppendix)
			sn := NewSecNum(number)
			if h.FirstChild() != nil {
				h.InsertBefore(h, h.FirstChild(), sn)
			} else {
				h.AppendChild(h, sn)
			}
		}
		if intoc {
			headings = append(headings, HeadingEntry{
				Level:  h.Level,
				Number: number,
				Title:  title,
				ID:     idOf(h),
			})
		}
		return gast.WalkSkipChildren, nil
	})

	// Pass 2: number citations by first appearance, build the reference list.
	citeNum := map[string]int{}
	var bib []BibEntry
	next := 0
	_ = gast.Walk(doc, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		c, ok := n.(*Citation)
		if !ok || !entering {
			return gast.WalkContinue, nil
		}
		ref, found := refByKey[c.Key]
		if !found {
			c.Resolved = false
			return gast.WalkContinue, nil
		}
		num, seen := citeNum[c.Key]
		if !seen {
			next++
			num = next
			citeNum[c.Key] = num
			bib = append(bib, BibEntry{Number: num, Key: c.Key, Ref: ref})
		}
		c.Number = num
		c.RefID = refID(c.Key)
		c.Resolved = true
		return gast.WalkContinue, nil
	})

	// Pass 3: hand the collected data to the directive nodes.
	_ = gast.Walk(doc, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		d, ok := n.(*Directive)
		if !ok || !entering {
			return gast.WalkContinue, nil
		}
		switch d.Name {
		case "toc":
			depth := tocDepth(d)
			for _, h := range headings {
				if h.Level <= depth {
					d.Headings = append(d.Headings, h)
				}
			}
		case "bibliography":
			d.Bib = bib
		}
		return gast.WalkSkipChildren, nil
	})
}

// wrapMatter replaces top-level `:::frontmatter` / `:::mainmatter` /
// `:::appendix` markers with Matter containers holding the nodes that follow
// each marker (up to the next marker). Content before the first marker is left
// in place, so documents that don't use matter markers are untouched.
func wrapMatter(doc *gast.Document) {
	var kids []gast.Node
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		kids = append(kids, c)
	}
	var container *Matter
	for _, n := range kids {
		if d, ok := n.(*Directive); ok {
			if region, isMarker := matterMarkers[d.Name]; isMarker {
				container = NewMatter(region)
				doc.InsertBefore(doc, n, container)
				doc.RemoveChild(doc, n)
				continue
			}
		}
		if container != nil {
			container.AppendChild(container, n) // also detaches n from doc
		}
	}
}

// matterOf returns the region ("front"/"main"/"appendix") of the nearest Matter
// ancestor, or "" when the node is outside any matter region.
func matterOf(n gast.Node) string {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if m, ok := p.(*Matter); ok {
			return m.Region
		}
	}
	return ""
}

func formatNumber(counters []int, level int, appendix bool) string {
	parts := make([]string, 0, level)
	for i := 1; i <= level; i++ {
		if appendix && i == 1 {
			parts = append(parts, string(rune('A'+counters[1]-1)))
		} else {
			parts = append(parts, strconv.Itoa(counters[i]))
		}
	}
	return strings.Join(parts, ".")
}

func tocDepth(d *Directive) int {
	if v := d.Options["depth"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultTOCDepth
}

// refID is the element id shared by a bibliography entry and the citations that
// link to it.
func refID(key string) string {
	var b strings.Builder
	b.WriteString("mdoc-ref-")
	for _, r := range strings.ToLower(key) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		case r == '-' || r == '_' || r == ' ':
			b.WriteByte('-')
		}
	}
	return b.String()
}

func classesOf(n gast.Node) []string {
	v, ok := n.AttributeString("class")
	if !ok {
		return nil
	}
	if b, ok := v.([]byte); ok {
		return strings.Fields(string(b))
	}
	return nil
}

func idOf(n gast.Node) string {
	v, ok := n.AttributeString("id")
	if !ok {
		return ""
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return ""
}

func nodeText(n gast.Node, source []byte) string {
	var b strings.Builder
	_ = gast.Walk(n, func(c gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *gast.Text:
			b.Write(t.Segment.Value(source))
		case *gast.String:
			b.Write(t.Value)
		}
		return gast.WalkContinue, nil
	})
	return b.String()
}
