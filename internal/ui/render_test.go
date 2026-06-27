package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
)

// plain strips ANSI styling so assertions on rendered text are stable. Lip
// Gloss v2 always embeds color escapes at render time (the profile downgrade
// happens at the terminal writer), so tests strip them here.
func plain(s string) string { return ansi.Strip(s) }

func sampleFile() diff.File {
	return diff.File{
		Path:    "hello.go",
		Status:  diff.StatusModified,
		Added:   1,
		Removed: 1,
		Hunks: []diff.Hunk{{
			Header:   "@@ -1,3 +1,3 @@",
			OldStart: 1, OldLines: 3, NewStart: 1, NewLines: 3,
			Lines: []diff.Line{
				{Type: diff.LineContext, Content: "package main", OldNumber: 1, NewNumber: 1},
				{Type: diff.LineRemoved, Content: "func old() {}", OldNumber: 2},
				{Type: diff.LineAdded, Content: "func new() {}", NewNumber: 2},
				{Type: diff.LineContext, Content: "// trailing", OldNumber: 3, NewNumber: 3},
			},
		}},
	}
}

func TestRenderInline_ShowsMarkersContentAndNumbers(t *testing.T) {
	out := RenderInline(sampleFile(), RenderOpts{Width: 80})
	lines := strings.Split(out, "\n")

	if !containsLine(lines, func(l string) bool {
		return strings.Contains(l, "-") && strings.Contains(l, "func old() {}") && strings.Contains(l, "2")
	}) {
		t.Errorf("removed line not rendered with marker/number:\n%s", out)
	}
	if !containsLine(lines, func(l string) bool {
		return strings.Contains(l, "+") && strings.Contains(l, "func new() {}")
	}) {
		t.Errorf("added line not rendered with marker:\n%s", out)
	}
	if !containsLine(lines, func(l string) bool {
		return strings.Contains(l, "package main")
	}) {
		t.Errorf("context line not rendered:\n%s", out)
	}
}

func containsLine(lines []string, pred func(string) bool) bool {
	for _, l := range lines {
		if pred(l) {
			return true
		}
	}
	return false
}

func TestRenderInline_VisualizeWhitespace(t *testing.T) {
	f := diff.File{Path: "ws.go", Hunks: []diff.Hunk{{
		Header: "@@ -1,1 +1,1 @@",
		Lines: []diff.Line{
			{Type: diff.LineAdded, Content: "\tx = 1", NewNumber: 1},
		},
	}}}
	out := RenderInline(f, RenderOpts{Width: 80, ShowWhitespace: true})
	if !strings.Contains(out, "→") {
		t.Errorf("tab not visualized as arrow:\n%s", out)
	}
	if !strings.Contains(out, "·") {
		t.Errorf("space not visualized as dot:\n%s", out)
	}

	// When ShowWhitespace is off, raw characters are preserved.
	out = RenderInline(f, RenderOpts{Width: 80, ShowWhitespace: false})
	if strings.Contains(out, "→") || strings.Contains(out, "·") {
		t.Errorf("whitespace visualized when toggle off:\n%s", out)
	}
}

func TestRenderInline_WhitespaceOnlyHighlighted(t *testing.T) {
	wsLine := diff.Line{Type: diff.LineAdded, Content: "  x", NewNumber: 1, WhitespaceOnly: true}
	normalLine := diff.Line{Type: diff.LineAdded, Content: "  x", NewNumber: 1}
	wsOut := renderInlineLine(wsLine, RenderOpts{Width: 80})
	normalOut := renderInlineLine(normalLine, RenderOpts{Width: 80})
	if wsOut == normalOut {
		t.Errorf("whitespace-only line should be styled differently from a normal added line")
	}
}

func TestRenderInline_BinaryMessageWithSizes(t *testing.T) {
	f := diff.File{Path: "logo.png", Status: diff.StatusModified, IsBinary: true,
		BinarySizeBefore: 24 * 1024, BinarySizeAfter: 31 * 1024}
	out := RenderInline(f, RenderOpts{Width: 80})
	if !strings.Contains(out, "Binary file: diff not available") {
		t.Errorf("missing binary message:\n%s", out)
	}
	if !strings.Contains(out, "24KB") || !strings.Contains(out, "31KB") {
		t.Errorf("missing sizes:\n%s", out)
	}
}

func TestRenderSideBySide_PairsChangedLines(t *testing.T) {
	out := RenderSideBySide(sampleFile(), RenderOpts{Width: 100})
	// The removed/added pair should appear on the same visual row: old on the
	// left of the column separator, new on the right.
	found := false
	for _, row := range strings.Split(out, "\n") {
		l, r, ok := strings.Cut(row, "│")
		if !ok {
			continue
		}
		if strings.Contains(l, "func old() {}") && strings.Contains(r, "func new() {}") {
			found = true
		}
	}
	if !found {
		t.Errorf("changed lines not paired across columns:\n%s", out)
	}
}

func TestRenderSideBySide_AddedFileEmptyLeft(t *testing.T) {
	f := diff.File{Path: "new.txt", Status: diff.StatusAdded, Hunks: []diff.Hunk{{
		Header: "@@ -0,0 +1,2 @@",
		Lines: []diff.Line{
			{Type: diff.LineAdded, Content: "hello", NewNumber: 1},
			{Type: diff.LineAdded, Content: "world", NewNumber: 2},
		},
	}}}
	out := RenderSideBySide(f, RenderOpts{Width: 100})
	for _, row := range strings.Split(out, "\n") {
		l, r, ok := strings.Cut(row, "│")
		if !ok {
			continue
		}
		if strings.ContainsAny(plain(l), "abcdefghijklmnopqrstuvwxyz") {
			t.Errorf("added file should have empty left column, got left=%q", plain(l))
		}
		_ = r
	}
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("added content missing from right column:\n%s", out)
	}
}

func TestRenderSideBySide_DeletedFileEmptyRight(t *testing.T) {
	f := diff.File{Path: "gone.txt", Status: diff.StatusDeleted, Hunks: []diff.Hunk{{
		Header: "@@ -1,2 +0,0 @@",
		Lines: []diff.Line{
			{Type: diff.LineRemoved, Content: "hello", OldNumber: 1},
			{Type: diff.LineRemoved, Content: "world", OldNumber: 2},
		},
	}}}
	out := RenderSideBySide(f, RenderOpts{Width: 100})
	for _, row := range strings.Split(out, "\n") {
		_, r, ok := strings.Cut(row, "│")
		if !ok {
			continue
		}
		if strings.ContainsAny(plain(r), "abcdefghijklmnopqrstuvwxyz") {
			t.Errorf("deleted file should have empty right column, got right=%q", plain(r))
		}
	}
}

func TestRenderInline_SearchHighlight(t *testing.T) {
	f := diff.File{Path: "s.go", Hunks: []diff.Hunk{{
		Header: "@@ -1,1 +1,1 @@",
		Lines:  []diff.Line{{Type: diff.LineContext, Content: "foo bar foo", OldNumber: 1, NewNumber: 1}},
	}}}
	hit := RenderInline(f, RenderOpts{Width: 80, Search: "foo"})
	miss := RenderInline(f, RenderOpts{Width: 80})
	if hit == miss {
		t.Errorf("search matches should be visually highlighted")
	}
	// Content text is preserved regardless of highlighting.
	if !strings.Contains(plain(hit), "foo bar foo") {
		t.Errorf("search highlight altered content text:\n%q", plain(hit))
	}
}

func TestRenderSidebar_StatusCountsAndSelection(t *testing.T) {
	files := []diff.File{
		{Path: "a.go", Status: diff.StatusModified, Added: 3, Removed: 1},
		{Path: "b.go", Status: diff.StatusAdded, Added: 10},
		{Path: "c.go", Status: diff.StatusDeleted, Removed: 5},
	}
	out := RenderSidebar(files, 0, "", true)
	if !strings.Contains(out, "a.go") || !strings.Contains(out, "b.go") || !strings.Contains(out, "c.go") {
		t.Errorf("sidebar missing file names:\n%s", out)
	}
	if !strings.Contains(out, "M") || !strings.Contains(out, "A") || !strings.Contains(out, "D") {
		t.Errorf("sidebar missing status letters:\n%s", out)
	}
	if !strings.Contains(out, "+3") || !strings.Contains(out, "-1") {
		t.Errorf("sidebar missing change counts:\n%s", out)
	}
}

func TestRenderSidebar_Filter(t *testing.T) {
	files := []diff.File{
		{Path: "main.go", Status: diff.StatusModified},
		{Path: "readme.md", Status: diff.StatusModified},
	}
	out := RenderSidebar(files, 0, "main", true)
	if !strings.Contains(out, "main.go") {
		t.Errorf("filtered-in file missing:\n%s", out)
	}
	if strings.Contains(out, "readme.md") {
		t.Errorf("filtered-out file should not appear:\n%s", out)
	}
}

func TestRenderControlBar_ShowsState(t *testing.T) {
	out := RenderControlBar("staged", "side-by-side", true, focusScope)
	for _, want := range []string{"staged", "side-by-side"} {
		if !strings.Contains(out, want) {
			t.Errorf("control bar missing %q:\n%s", want, out)
		}
	}
}

func TestRenderEmptyMessage(t *testing.T) {
	if !strings.Contains(RenderEmptyMessage(), "working tree is clean") {
		t.Errorf("empty message wrong: %q", RenderEmptyMessage())
	}
}

func TestRenderFooter_ShowsKeyHints(t *testing.T) {
	out := RenderFooter()
	if !strings.Contains(out, "quit") {
		t.Errorf("footer should hint quit: %q", out)
	}
}
