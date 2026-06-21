package mdext

import (
	"slices"
	"strconv"
	"strings"

	"github.com/hinkolas/mdoc/internal/document"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const defaultTOCDepth = 3

// transformer is the foundation pass: it numbers headings and captioned
// figures/tables, builds their captions, resolves `[#id]` cross-references,
// numbers `[@key]` citations by first appearance, and attaches the collected
// lists onto the directive nodes (renderers get no parser.Context, so data must
// travel on the nodes).
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

	// Pass 1: walk the document in order, numbering headings and captioned
	// figures/tables. A heading's region sets the defaults (front = unnumbered +
	// out of the TOC; main = decimal; appendix = lettered); per-heading classes
	// override them. Figures/tables number per chapter (2.1, A.1), restarting at
	// each top-level heading and falling back to a continuous count when there is
	// no chapter number. Every numbered element is recorded for cross-references.
	var headings []HeadingEntry
	var figures, tables []CaptionEntry
	xrefNum := map[string]string{}   // id -> number (may be "")
	xrefTitle := map[string]string{} // id -> plain title (for numberless targets)
	ids := map[string]bool{}         // ids that exist (for page references)
	counters := make([]int, 7)       // indices 1..6
	prevAppendix := false
	chapter := ""                // current top-level heading number
	chapFig, chapTab := 0, 0     // per-chapter figure/table counters
	globalFig, globalTab := 0, 0 // counters used when there is no chapter
	_ = gast.Walk(doc, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *gast.Heading:
			classes := classesOf(node)
			region := matterOf(node)

			numbered := levelNumbered(t.cfg.Numbering, node.Level)
			intoc := true
			if region == "front" {
				numbered, intoc = false, false
			}
			if slices.Contains(classes, "unnumbered") {
				numbered = false
			} else if slices.Contains(classes, "numbered") {
				numbered = levelNumbered(t.cfg.Numbering, node.Level)
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

			title := nodeText(node, source) // before injecting the number
			number := ""
			if numbered && node.Level >= 1 && node.Level < len(counters) {
				counters[node.Level]++
				for i := node.Level + 1; i < len(counters); i++ {
					counters[i] = 0
				}
				number = renderNumber(t.cfg.Numbering, counters, node.Level, isAppendix)
				sn := NewSecNum(number)
				if node.FirstChild() != nil {
					node.InsertBefore(node, node.FirstChild(), sn)
				} else {
					node.AppendChild(node, sn)
				}
			}

			id := idOf(node)
			if id != "" {
				ids[id] = true
				xrefNum[id] = number
				xrefTitle[id] = title
			}
			if node.Level == 1 {
				chapter, chapFig, chapTab = number, 0, 0
			}
			if intoc {
				headings = append(headings, HeadingEntry{
					Level: node.Level, Number: number, Title: title, ID: id,
				})
			}
			return gast.WalkSkipChildren, nil

		case *Captioned:
			var number string
			isTable := node.Variant == "table"
			switch {
			case isTable && chapter == "":
				globalTab++
				number = strconv.Itoa(globalTab)
			case isTable:
				chapTab++
				number = chapter + "." + strconv.Itoa(chapTab)
			case chapter == "":
				globalFig++
				number = strconv.Itoa(globalFig)
			default:
				chapFig++
				number = chapter + "." + strconv.Itoa(chapFig)
			}
			node.Number = number
			if node.ID == "" {
				node.ID = captionID(node.Variant, number)
			}

			title := buildCaption(node, source, t.cfg.label(node.Variant)+" "+number, labelClass(node.Variant))
			entry := CaptionEntry{Number: number, Title: title, ID: node.ID}
			if isTable {
				tables = append(tables, entry)
			} else {
				figures = append(figures, entry)
			}
			ids[node.ID] = true
			xrefNum[node.ID] = number
			return gast.WalkSkipChildren, nil
		}
		return gast.WalkContinue, nil
	})

	// Pass 1b: resolve `[#id]` cross-references against the collected numbers.
	// "num" shows the target's number (or its title when it has none); "page"
	// emits a link the theme resolves to a page number via target-counter.
	_ = gast.Walk(doc, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		x, ok := n.(*Xref)
		if !ok || !entering {
			return gast.WalkContinue, nil
		}
		switch x.Mode {
		case "page":
			x.Resolved = ids[x.ID]
		default:
			if num, ok := xrefNum[x.ID]; ok {
				switch {
				case num != "":
					x.Number, x.Resolved = num, true
				case xrefTitle[x.ID] != "":
					x.Number, x.Resolved = xrefTitle[x.ID], true
				}
			}
		}
		return gast.WalkContinue, nil
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
		case "lof":
			d.Entries = figures
		case "lot":
			d.Entries = tables
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
	// The AST keeps text raw, so a caption written with "&nbsp;" yields a literal
	// entity here. Decode references so plain-text titles (TOC / LOF / LOT) read
	// the way the rendered caption does, not "50&nbsp;Hz".
	return decodeEntities(b.String())
}

// decodeEntities replaces HTML character references (&name; / &#dd; / &#xhh;)
// in s with their characters, leaving everything else untouched.
func decodeEntities(s string) string {
	if !strings.ContainsRune(s, '&') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '&' {
			b.WriteByte(s[i])
			i++
			continue
		}
		if semi := strings.IndexByte(s[i:], ';'); semi > 1 && semi < 32 {
			if dec, ok := decodeRef(s[i+1 : i+semi]); ok {
				b.WriteString(dec)
				i += semi + 1
				continue
			}
		}
		b.WriteByte('&')
		i++
	}
	return b.String()
}

func decodeRef(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}
	if ref[0] == '#' {
		base, digits := 10, ref[1:]
		if len(ref) > 1 && (ref[1] == 'x' || ref[1] == 'X') {
			base, digits = 16, ref[2:]
		}
		v, err := strconv.ParseUint(digits, base, 32)
		if err != nil || v == 0 {
			return "", false
		}
		return string(rune(v)), true
	}
	if e, ok := util.LookUpHTML5EntityByName(ref); ok {
		return string(e.Characters), true
	}
	return "", false
}

// labelClass is the CSS class of a captioned variant's injected label.
func labelClass(variant string) string {
	if variant == "table" {
		return "mdoc-tab-label"
	}
	return "mdoc-fig-label"
}

// captionID synthesises an element id for a captioned block that the author
// didn't label, e.g. ("figure", "2.1") -> "fig-2-1".
func captionID(variant, number string) string {
	prefix := "fig"
	if variant == "table" {
		prefix = "tab"
	}
	return prefix + "-" + strings.ReplaceAll(strings.ToLower(number), ".", "-")
}

// buildCaption separates a captioned block's media from its caption: image-only
// paragraphs (and any non-paragraph block, e.g. a table) stay as the figure
// media, while the remaining text paragraphs are folded into a Caption node led
// by the injected label. The caption goes below the media for figures and above
// it for tables. It returns the plain caption text for the list of figures/
// tables (falling back to the first image's alt text).
func buildCaption(cap *Captioned, source []byte, label, class string) string {
	var captionParas []*gast.Paragraph
	for c := cap.FirstChild(); c != nil; c = c.NextSibling() {
		if p, ok := c.(*gast.Paragraph); ok && !isImageOnlyParagraph(p, source) {
			captionParas = append(captionParas, p)
		}
	}

	title := ""
	for _, p := range captionParas {
		title += nodeText(p, source)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = strings.TrimSpace(firstImageAlt(cap, source))
	}

	capNode := NewCaption()
	capNode.AppendChild(capNode, NewCaptionLabel(label, class))
	for i, p := range captionParas {
		if i > 0 {
			capNode.AppendChild(capNode, gast.NewString([]byte(" ")))
		}
		for ch := p.FirstChild(); ch != nil; {
			next := ch.NextSibling()
			capNode.AppendChild(capNode, ch) // re-parents (detaches from p)
			ch = next
		}
		cap.RemoveChild(cap, p)
	}

	if cap.Variant == "table" {
		cap.InsertBefore(cap, cap.FirstChild(), capNode)
	} else {
		cap.AppendChild(cap, capNode)
	}
	return title
}

// isImageOnlyParagraph reports whether a paragraph holds only images (plus
// whitespace) — the heuristic that classifies a paragraph as figure media
// rather than caption text.
func isImageOnlyParagraph(p *gast.Paragraph, source []byte) bool {
	hasImage := false
	for c := p.FirstChild(); c != nil; c = c.NextSibling() {
		switch v := c.(type) {
		case *gast.Image:
			hasImage = true
		case *gast.Text:
			if len(strings.TrimSpace(string(v.Segment.Value(source)))) != 0 {
				return false
			}
		default:
			return false
		}
	}
	return hasImage
}

// firstImageAlt returns the alt text of the first image under n, or "".
func firstImageAlt(n gast.Node, source []byte) string {
	alt := ""
	_ = gast.Walk(n, func(c gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering || alt != "" {
			return gast.WalkContinue, nil
		}
		if img, ok := c.(*gast.Image); ok {
			alt = nodeText(img, source)
			return gast.WalkStop, nil
		}
		return gast.WalkContinue, nil
	})
	return alt
}
