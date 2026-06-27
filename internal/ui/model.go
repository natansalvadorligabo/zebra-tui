package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
	"github.com/salvadorligabo/zebra-tui/internal/git"
)

type focusPanel int

const (
	focusSidebar focusPanel = iota
	focusScope
	focusView
	focusWhitespace
	focusDiff
)

// focusOrder is the Tab cycle order through focusable elements.
var focusOrder = []focusPanel{focusSidebar, focusScope, focusView, focusWhitespace, focusDiff}

type viewMode int

const (
	viewInline viewMode = iota
	viewSideBySide
)

func (v viewMode) String() string {
	if v == viewSideBySide {
		return "side-by-side"
	}
	return "inline"
}

// Model is the root Bubble Tea model for zebra.
type Model struct {
	repo  string
	scope git.Scope
	files []diff.File

	selected int // sidebar cursor: index into the filtered file list
	opened   int // index (into filtered list) of the file shown in the diff panel

	focus          focusPanel
	view           viewMode
	showWhitespace bool

	width, height int

	// sidebar filter state
	filterActive bool
	filter       string

	// diff search state
	searchActive bool
	search       string
	matches      []int
	matchIdx     int

	diffScroll int

	loaded  bool
	loadErr error
}

// New creates a root model that loads its files on Init.
func New(repo string, scope git.Scope) Model {
	return Model{repo: repo, scope: scope}
}

// NewWithFiles creates a root model pre-seeded with already-loaded files so the
// first frame can render without a round-trip to git.
func NewWithFiles(repo string, scope git.Scope, files []diff.File) Model {
	return Model{repo: repo, scope: scope, files: files, loaded: true}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.loaded {
		return nil
	}
	return loadFilesCmd(m.repo, m.scope)
}

// focusedFiles returns the currently visible (filtered) file list.
func (m Model) filteredFiles() []diff.File {
	return FilterFiles(m.files, m.filter)
}

func (m *Model) cycleFocus(forward bool) {
	idx := 0
	for i, f := range focusOrder {
		if f == m.focus {
			idx = i
			break
		}
	}
	if forward {
		idx = (idx + 1) % len(focusOrder)
	} else {
		idx = (idx - 1 + len(focusOrder)) % len(focusOrder)
	}
	m.focus = focusOrder[idx]
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case filesLoadedMsg:
		m.files = msg.files
		m.loadErr = msg.err
		m.loaded = true
		if m.selected >= len(m.filteredFiles()) {
			m.selected = 0
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Ctrl+C always quits, even while typing in an input.
	if key == "ctrl+c" {
		return m, tea.Quit
	}
	// Input modes capture text before normal keybindings apply.
	if m.filterActive {
		return m.handleFilterKey(key, msg.Text), nil
	}
	if m.searchActive {
		return m.handleSearchKey(key, msg.Text), nil
	}

	if key == "ctrl+f" {
		if m.focus == focusSidebar {
			m.filterActive = true
		} else {
			m.searchActive = true
			m.recomputeMatches()
		}
		return m, nil
	}

	switch key {
	case "q":
		return m, tea.Quit
	case "tab":
		m.cycleFocus(true)
		return m, nil
	case "shift+tab":
		m.cycleFocus(false)
		return m, nil
	case "left":
		m.focus = focusSidebar
		return m, nil
	case "right":
		m.focus = focusDiff
		return m, nil
	case "up", "k":
		return m.moveUp(), nil
	case "down", "j":
		return m.moveDown(), nil
	case "space":
		return m.activateFocused()
	case "enter":
		if m.focus == focusSidebar {
			m.opened = m.selected
			return m, nil
		}
		return m.activateFocused()
	case "n":
		m.diffScroll = m.nextHunkRow()
		return m, nil
	case "p":
		m.diffScroll = m.prevHunkRow()
		return m, nil
	case "s":
		m.scope = cycleScope(m.scope)
		m.diffScroll = 0
		return m, loadFilesCmd(m.repo, m.scope)
	case "v":
		m.view = toggleView(m.view)
		return m, nil
	case "w":
		m.showWhitespace = !m.showWhitespace
		return m, nil
	}
	return m, nil
}

// handleFilterKey routes keys while the sidebar filter input is active. key is
// the keystroke string; text is the printable characters (empty for non-text
// keys), used to build the query.
func (m Model) handleFilterKey(key, text string) Model {
	switch key {
	case "esc":
		m.filterActive = false
		m.filter = ""
		m.selected = 0
	case "enter":
		m.filterActive = false // accept filter, keep it applied
	case "backspace":
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
		}
		m.selected = 0
	default:
		if text != "" {
			m.filter += text
			m.selected = 0
		}
	}
	return m
}

// handleSearchKey routes keys while the diff search input is active.
func (m Model) handleSearchKey(key, text string) Model {
	switch key {
	case "esc":
		m.searchActive = false
		m.search = ""
		m.matches = nil
		m.matchIdx = 0
	case "backspace":
		if r := []rune(m.search); len(r) > 0 {
			m.search = string(r[:len(r)-1])
		}
		m.recomputeMatches()
	case "up":
		m.gotoMatch(-1)
	case "down", "enter":
		m.gotoMatch(1)
	default:
		if text != "" {
			m.search += text
			m.recomputeMatches()
		}
	}
	return m
}

// recomputeMatches finds rendered rows containing the search query and scrolls
// to the first one.
func (m *Model) recomputeMatches() {
	m.matches = nil
	m.matchIdx = 0
	if m.search == "" {
		return
	}
	for i, text := range m.renderedRowText() {
		if strings.Contains(text, m.search) {
			m.matches = append(m.matches, i)
		}
	}
	if len(m.matches) > 0 {
		m.diffScroll = m.matches[0]
	}
}

// gotoMatch advances the current match by dir (+1 next, -1 previous), wrapping.
func (m *Model) gotoMatch(dir int) {
	if len(m.matches) == 0 {
		return
	}
	m.matchIdx = (m.matchIdx + dir + len(m.matches)) % len(m.matches)
	m.diffScroll = m.matches[m.matchIdx]
}

// renderedRowText returns the textual content of each rendered diff row of the
// opened file, in render order, so match rows align with scroll offsets.
func (m Model) renderedRowText() []string {
	files := m.filteredFiles()
	if len(files) == 0 || m.opened < 0 || m.opened >= len(files) {
		return nil
	}
	f := files[m.opened]
	var rows []string
	for i, h := range f.Hunks {
		if i > 0 {
			rows = append(rows, "")
		}
		rows = append(rows, h.Header)
		if m.view == viewSideBySide {
			for _, pr := range pairRows(h.Lines) {
				var l, r string
				if pr.left != nil {
					l = pr.left.Content
				}
				if pr.right != nil {
					r = pr.right.Content
				}
				rows = append(rows, l+" "+r)
			}
		} else {
			for _, ln := range h.Lines {
				rows = append(rows, ln.Content)
			}
		}
	}
	return rows
}

func toggleView(v viewMode) viewMode {
	if v == viewInline {
		return viewSideBySide
	}
	return viewInline
}

// activateFocused performs the action of the focused control (Space/Enter).
func (m Model) activateFocused() (tea.Model, tea.Cmd) {
	switch m.focus {
	case focusScope:
		m.scope = cycleScope(m.scope)
		m.diffScroll = 0
		return m, loadFilesCmd(m.repo, m.scope)
	case focusView:
		m.view = toggleView(m.view)
	case focusWhitespace:
		m.showWhitespace = !m.showWhitespace
	}
	return m, nil
}

func cycleScope(s git.Scope) git.Scope {
	switch s {
	case git.ScopeWorkingTree:
		return git.ScopeStaged
	case git.ScopeStaged:
		return git.ScopeAll
	default:
		return git.ScopeWorkingTree
	}
}

// hunkStartRows returns the rendered row offset at which each hunk of the
// opened file begins, matching the inline/side-by-side render layout.
func (m Model) hunkStartRows() []int {
	files := m.filteredFiles()
	if len(files) == 0 || m.opened < 0 || m.opened >= len(files) {
		return nil
	}
	f := files[m.opened]
	var rows []int
	row := 0
	for i, h := range f.Hunks {
		if i > 0 {
			row++ // blank separator line between hunks
		}
		rows = append(rows, row)
		lineCount := len(h.Lines)
		if m.view == viewSideBySide {
			lineCount = len(pairRows(h.Lines))
		}
		row += 1 + lineCount // header + body lines
	}
	return rows
}

// nextHunkRow returns the scroll offset of the first hunk starting after the
// current scroll position, or the last hunk if already past it.
func (m Model) nextHunkRow() int {
	rows := m.hunkStartRows()
	for _, r := range rows {
		if r > m.diffScroll {
			return r
		}
	}
	if len(rows) > 0 {
		return rows[len(rows)-1]
	}
	return m.diffScroll
}

// prevHunkRow returns the scroll offset of the last hunk starting before the
// current scroll position, or the first hunk.
func (m Model) prevHunkRow() int {
	rows := m.hunkStartRows()
	prev := 0
	for _, r := range rows {
		if r < m.diffScroll {
			prev = r
		}
	}
	return prev
}

// moveUp moves the sidebar cursor up or scrolls the diff up, per focus.
func (m Model) moveUp() Model {
	if m.focus == focusDiff {
		if m.diffScroll > 0 {
			m.diffScroll--
		}
		return m
	}
	if m.selected > 0 {
		m.selected--
	}
	return m
}

// moveDown moves the sidebar cursor down or scrolls the diff down, per focus.
func (m Model) moveDown() Model {
	if m.focus == focusDiff {
		if max := len(m.renderedRowText()) - 1; m.diffScroll < max {
			m.diffScroll++
		}
		return m
	}
	if m.selected < len(m.filteredFiles())-1 {
		m.selected++
	}
	return m
}
