package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
)

// lipglossWidth returns the rendered (visible) width of s, ignoring ANSI codes.
func lipglossWidth(s string) int { return lipgloss.Width(s) }

// RenderOpts controls how a file's diff is rendered.
type RenderOpts struct {
	Width          int
	ShowWhitespace bool
	Search         string // when non-empty, matches are highlighted
}

const (
	gutterNumWidth = 4
	binaryMessage  = "Binary file: diff not available"
	emptyMessage   = "Nothing to diff: working tree is clean"
)

// RenderInline renders a file's diff in single-column (inline) mode.
func RenderInline(f diff.File, opts RenderOpts) string {
	if f.IsBinary {
		return renderBinary(f)
	}
	var b strings.Builder
	for hi, h := range f.Hunks {
		if hi > 0 {
			b.WriteString("\n")
		}
		b.WriteString(styleLineNumber.Render(h.Header))
		b.WriteString("\n")
		for _, ln := range h.Lines {
			b.WriteString(renderInlineLine(ln, opts))
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderInlineLine(ln diff.Line, opts RenderOpts) string {
	marker := " "
	style := styleContext
	switch ln.Type {
	case diff.LineAdded:
		marker = "+"
		style = styleAdded
	case diff.LineRemoved:
		marker = "-"
		style = styleRemoved
	}
	if ln.WhitespaceOnly {
		style = styleWhitespaceOnly
	}
	gutter := styleLineNumber.Render(fmt.Sprintf("%*s %*s", gutterNumWidth, num(ln.OldNumber), gutterNumWidth, num(ln.NewNumber)))
	content := transformContent(ln.Content, opts)
	return gutter + " " + style.Render(marker+" "+content)
}

// num formats a line number, blank when zero.
func num(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

// transformContent applies whitespace visualization and search highlighting.
func transformContent(s string, opts RenderOpts) string {
	if opts.ShowWhitespace {
		s = visualizeWhitespace(s)
	}
	if opts.Search != "" {
		s = highlightMatches(s, opts.Search)
	}
	return s
}

func visualizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\t", "→")
	s = strings.ReplaceAll(s, " ", "·")
	return s
}

func highlightMatches(s, query string) string {
	if query == "" {
		return s
	}
	var b strings.Builder
	rest := s
	for {
		idx := strings.Index(rest, query)
		if idx < 0 {
			b.WriteString(rest)
			break
		}
		b.WriteString(rest[:idx])
		b.WriteString(styleSearchMatch.Render(query))
		rest = rest[idx+len(query):]
	}
	return b.String()
}

// sideRow holds the optional old (left) and new (right) line for one visual row.
type sideRow struct {
	left  *diff.Line
	right *diff.Line
}

// RenderSideBySide renders a file's diff in two columns: original on the left,
// new on the right, each with independent line numbers.
func RenderSideBySide(f diff.File, opts RenderOpts) string {
	if f.IsBinary {
		return renderBinary(f)
	}
	colWidth := (opts.Width - 3) / 2
	if colWidth < 10 {
		colWidth = 10
	}
	var b strings.Builder
	for hi, h := range f.Hunks {
		if hi > 0 {
			b.WriteString("\n")
		}
		b.WriteString(styleLineNumber.Render(h.Header))
		b.WriteString("\n")
		for _, row := range pairRows(h.Lines) {
			left := renderSideCell(row.left, true, opts)
			right := renderSideCell(row.right, false, opts)
			b.WriteString(padTo(left, colWidth))
			b.WriteString(" │ ")
			b.WriteString(right)
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// pairRows aligns removed lines (left) with the added lines that follow (right)
// so a modification shows old and new on the same row.
func pairRows(lines []diff.Line) []sideRow {
	var rows []sideRow
	i := 0
	for i < len(lines) {
		ln := lines[i]
		if ln.Type == diff.LineContext {
			l := lines[i]
			rows = append(rows, sideRow{left: &l, right: &l})
			i++
			continue
		}
		// Gather a run of removed then a run of added.
		remStart := i
		for i < len(lines) && lines[i].Type == diff.LineRemoved {
			i++
		}
		removed := lines[remStart:i]
		addStart := i
		for i < len(lines) && lines[i].Type == diff.LineAdded {
			i++
		}
		added := lines[addStart:i]

		n := len(removed)
		if len(added) > n {
			n = len(added)
		}
		for k := 0; k < n; k++ {
			var row sideRow
			if k < len(removed) {
				r := removed[k]
				row.left = &r
			}
			if k < len(added) {
				a := added[k]
				row.right = &a
			}
			rows = append(rows, row)
		}
	}
	return rows
}

// renderSideCell renders one column cell. left=true uses the old line number;
// otherwise the new line number. A nil line yields an empty (blank) cell.
func renderSideCell(ln *diff.Line, left bool, opts RenderOpts) string {
	if ln == nil {
		return styleLineNumber.Render(strings.Repeat(" ", gutterNumWidth)) + "   "
	}
	number := ln.NewNumber
	marker := " "
	style := styleContext
	switch ln.Type {
	case diff.LineAdded:
		marker, style = "+", styleAdded
	case diff.LineRemoved:
		marker, style = "-", styleRemoved
	}
	if left {
		number = ln.OldNumber
	}
	if ln.WhitespaceOnly {
		style = styleWhitespaceOnly
	}
	gutter := styleLineNumber.Render(fmt.Sprintf("%*s", gutterNumWidth, num(number)))
	content := transformContent(ln.Content, opts)
	return gutter + " " + style.Render(marker+" "+content)
}

// padTo pads s with spaces to the given visible width (truncating if longer).
func padTo(s string, w int) string {
	vis := lipglossWidth(s)
	if vis >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

func renderBinary(f diff.File) string {
	msg := binaryMessage
	if f.BinarySizeBefore > 0 || f.BinarySizeAfter > 0 {
		msg += fmt.Sprintf("  (%s → %s)", humanSize(f.BinarySizeBefore), humanSize(f.BinarySizeAfter))
	}
	return styleMessage.Render(msg)
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%cB", float64(n)/float64(div), "KMGT"[exp])
}
