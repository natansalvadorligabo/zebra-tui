package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
	"github.com/salvadorligabo/zebra-tui/internal/git"
)

// key builds a v2 key-press message matching what real terminal input produces:
// special keys carry a Code, printable runes additionally carry Text.
func key(s string) tea.KeyPressMsg {
	switch s {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "ctrl+f":
		return tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	}
	// A printable rune: Code is the rune, Text carries the character(s).
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

// send applies a sequence of keys and returns the resulting model.
func send(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

func twoFileModel() Model {
	files := []diff.File{
		{Path: "a.go", Status: diff.StatusModified, Added: 2, Removed: 1, Hunks: []diff.Hunk{{
			Header: "@@ -1,2 +1,2 @@",
			Lines: []diff.Line{
				{Type: diff.LineContext, Content: "x", OldNumber: 1, NewNumber: 1},
				{Type: diff.LineRemoved, Content: "old", OldNumber: 2},
				{Type: diff.LineAdded, Content: "new", NewNumber: 2},
			},
		}}},
		{Path: "b.go", Status: diff.StatusAdded, Added: 1},
	}
	m := New(".", git.ScopeWorkingTree)
	m.files = files
	return m
}

func TestUpdate_TabCyclesFocus(t *testing.T) {
	m := twoFileModel()
	if m.focus != focusSidebar {
		t.Fatalf("initial focus = %v, want sidebar", m.focus)
	}
	// Tabbing through the whole order returns to the start.
	got := []focusPanel{}
	for i := 0; i < len(focusOrder); i++ {
		m = send(m, "tab")
		got = append(got, m.focus)
	}
	if m.focus != focusSidebar {
		t.Errorf("after full cycle focus = %v, want sidebar", m.focus)
	}
}

func TestUpdate_ArrowsMovePanelFocus(t *testing.T) {
	m := twoFileModel()
	m = send(m, "right")
	if m.focus != focusDiff {
		t.Errorf("right arrow: focus = %v, want diff", m.focus)
	}
	m = send(m, "left")
	if m.focus != focusSidebar {
		t.Errorf("left arrow: focus = %v, want sidebar", m.focus)
	}
}

func TestUpdate_SidebarSelectionMoves(t *testing.T) {
	m := twoFileModel()
	if m.selected != 0 {
		t.Fatalf("initial selected = %d, want 0", m.selected)
	}
	m = send(m, "down")
	if m.selected != 1 {
		t.Errorf("after down selected = %d, want 1", m.selected)
	}
	// Clamp at the bottom.
	m = send(m, "down")
	if m.selected != 1 {
		t.Errorf("selection should clamp at last file, got %d", m.selected)
	}
	m = send(m, "k") // vim up
	if m.selected != 0 {
		t.Errorf("after k selected = %d, want 0", m.selected)
	}
}

func TestUpdate_EnterOpensFileKeepingSidebarFocus(t *testing.T) {
	m := twoFileModel()
	m = send(m, "down", "enter")
	if m.opened != 1 {
		t.Errorf("opened = %d, want 1", m.opened)
	}
	if m.focus != focusSidebar {
		t.Errorf("focus moved away from sidebar after enter: %v", m.focus)
	}
}

func TestUpdate_QuitReturnsQuitCmd(t *testing.T) {
	m := twoFileModel()
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Fatalf("q produced no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("q should return tea.Quit")
	}
}

func twoHunkModel() Model {
	f := diff.File{Path: "m.go", Status: diff.StatusModified, Hunks: []diff.Hunk{
		{Header: "@@ -1,3 +1,3 @@", Lines: []diff.Line{
			{Type: diff.LineContext, Content: "a", OldNumber: 1, NewNumber: 1},
			{Type: diff.LineRemoved, Content: "b", OldNumber: 2},
			{Type: diff.LineAdded, Content: "B", NewNumber: 2},
		}},
		{Header: "@@ -10,3 +10,3 @@", Lines: []diff.Line{
			{Type: diff.LineContext, Content: "x", OldNumber: 10, NewNumber: 10},
			{Type: diff.LineRemoved, Content: "y", OldNumber: 11},
			{Type: diff.LineAdded, Content: "Y", NewNumber: 11},
		}},
	}}
	m := New(".", git.ScopeWorkingTree)
	m.files = []diff.File{f}
	return m
}

func TestUpdate_DiffScroll(t *testing.T) {
	m := twoHunkModel()
	m = send(m, "right") // focus diff
	m = send(m, "down", "down")
	if m.diffScroll != 2 {
		t.Errorf("diffScroll = %d, want 2", m.diffScroll)
	}
	m = send(m, "up")
	if m.diffScroll != 1 {
		t.Errorf("diffScroll = %d, want 1", m.diffScroll)
	}
}

func TestUpdate_HunkNavigation(t *testing.T) {
	m := twoHunkModel()
	// Inline layout: hunk0 header(0) + 3 lines(1..3), blank(4), hunk1 header(5).
	m = send(m, "n")
	if m.diffScroll != 5 {
		t.Errorf("after n diffScroll = %d, want 5 (second hunk header)", m.diffScroll)
	}
	m = send(m, "p")
	if m.diffScroll != 0 {
		t.Errorf("after p diffScroll = %d, want 0 (first hunk header)", m.diffScroll)
	}
}

func TestUpdate_ViewToggle(t *testing.T) {
	m := twoFileModel()
	if m.view != viewInline {
		t.Fatalf("initial view = %v, want inline", m.view)
	}
	m = send(m, "v")
	if m.view != viewSideBySide {
		t.Errorf("after v view = %v, want side-by-side", m.view)
	}
	m = send(m, "v")
	if m.view != viewInline {
		t.Errorf("after second v view = %v, want inline", m.view)
	}
}

func TestUpdate_WhitespaceToggle(t *testing.T) {
	m := twoFileModel()
	m = send(m, "w")
	if !m.showWhitespace {
		t.Errorf("after w showWhitespace = false, want true")
	}
	m = send(m, "w")
	if m.showWhitespace {
		t.Errorf("after second w showWhitespace = true, want false")
	}
}

func TestUpdate_ScopeCycleReloads(t *testing.T) {
	m := twoFileModel()
	// Focus the scope toggle, then activate it.
	m = send(m, "tab") // sidebar -> scope
	if m.focus != focusScope {
		t.Fatalf("focus = %v, want scope", m.focus)
	}
	next, cmd := m.Update(key("space"))
	m = next.(Model)
	if m.scope != git.ScopeStaged {
		t.Errorf("scope = %v, want staged", m.scope)
	}
	if cmd == nil {
		t.Errorf("scope change should trigger a reload command")
	}
	// Cycle through all three back to working tree.
	m = send(m, "space") // staged -> all
	if m.scope != git.ScopeAll {
		t.Errorf("scope = %v, want all", m.scope)
	}
	m = send(m, "space") // all -> working
	if m.scope != git.ScopeWorkingTree {
		t.Errorf("scope = %v, want working tree", m.scope)
	}
}

func TestUpdate_FocusedToggleActivatesView(t *testing.T) {
	m := twoFileModel()
	m = send(m, "tab", "tab") // sidebar -> scope -> view
	if m.focus != focusView {
		t.Fatalf("focus = %v, want view", m.focus)
	}
	m = send(m, "space")
	if m.view != viewSideBySide {
		t.Errorf("space on focused view toggle should switch to side-by-side")
	}
}

func TestUpdate_SidebarFilter(t *testing.T) {
	m := twoFileModel() // a.go, b.go
	m = send(m, "ctrl+f")
	if !m.filterActive {
		t.Fatalf("ctrl+f on sidebar should activate filter")
	}
	m = send(m, "b")
	if m.filter != "b" {
		t.Errorf("filter = %q, want b", m.filter)
	}
	if got := m.filteredFiles(); len(got) != 1 || got[0].Path != "b.go" {
		t.Errorf("filtered files = %+v, want only b.go", got)
	}
	m = send(m, "esc")
	if m.filterActive || m.filter != "" {
		t.Errorf("esc should close and clear the filter (active=%v filter=%q)", m.filterActive, m.filter)
	}
}

func searchModel() Model {
	f := diff.File{Path: "s.go", Status: diff.StatusModified, Hunks: []diff.Hunk{{
		Header: "@@ -1,4 +1,4 @@",
		Lines: []diff.Line{
			{Type: diff.LineContext, Content: "alpha", OldNumber: 1, NewNumber: 1},
			{Type: diff.LineAdded, Content: "needle one", NewNumber: 2},
			{Type: diff.LineContext, Content: "beta", OldNumber: 2, NewNumber: 3},
			{Type: diff.LineAdded, Content: "needle two", NewNumber: 4},
		},
	}}}
	m := New(".", git.ScopeWorkingTree)
	m.files = []diff.File{f}
	return m
}

func TestUpdate_DiffSearchHighlightsAndNavigates(t *testing.T) {
	m := searchModel()
	m = send(m, "right") // focus diff
	m = send(m, "ctrl+f")
	if !m.searchActive {
		t.Fatalf("ctrl+f on diff should activate search")
	}
	m = send(m, "n", "e", "e", "d", "l", "e")
	if m.search != "needle" {
		t.Fatalf("search = %q, want needle", m.search)
	}
	// Two matches: rendered rows for "needle one" (row 2) and "needle two" (row 4).
	// Layout: header(0), alpha(1), needle one(2), beta(3), needle two(4).
	if len(m.matches) != 2 {
		t.Fatalf("matches = %v, want 2", m.matches)
	}
	if m.diffScroll != m.matches[0] {
		t.Errorf("diffScroll = %d, want first match %d", m.diffScroll, m.matches[0])
	}
	m = send(m, "down") // next match
	if m.matchIdx != 1 || m.diffScroll != m.matches[1] {
		t.Errorf("after down matchIdx=%d scroll=%d, want 1 and %d", m.matchIdx, m.diffScroll, m.matches[1])
	}
	m = send(m, "esc")
	if m.searchActive || m.search != "" {
		t.Errorf("esc should close diff search")
	}
}

func TestUpdate_DiffScrollClampsAtEnd(t *testing.T) {
	m := twoHunkModel()
	m = send(m, "right") // focus diff
	// Total rendered rows for the opened file.
	rows := len(m.renderedRowText())
	for i := 0; i < rows+10; i++ {
		m = send(m, "down")
	}
	if m.diffScroll > rows-1 {
		t.Errorf("diffScroll = %d exceeded last row index %d", m.diffScroll, rows-1)
	}
}
