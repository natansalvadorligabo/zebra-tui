package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
)

const sidebarWidth = 30

// View implements tea.Model. The diff UI runs in the alternate screen buffer.
func (m Model) View() tea.View {
	return tea.View{Content: m.content(), AltScreen: true}
}

// content renders the full-screen UI as a string.
func (m Model) content() string {
	if m.loadErr != nil {
		return styleMessage.Render("error: " + m.loadErr.Error())
	}

	title := RenderTitleBar(m.width)
	control := RenderControlBar(m.scope.String(), m.view.String(), m.showWhitespace, m.focus)

	h := m.diffHeight()
	sidebar := padLinesTo(RenderSidebar(m.files, m.selected, m.filter, m.focus == focusSidebar), h)
	diff := padLinesTo(m.renderDiffPanel(), h)
	sidebarBox := panelStyle(m.focus == focusSidebar).Width(sidebarWidth).Render(sidebar)
	diffBox := panelStyle(m.focus == focusDiff).Width(m.diffWidth()).Render(diff)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, diffBox)

	parts := []string{title, control, body}
	if m.searchActive {
		parts = append(parts, styleFocused.Render("search: "+m.search))
	}
	parts = append(parts, RenderFooter())
	return strings.Join(parts, "\n")
}

// renderDiffPanel renders the diff for the currently opened file.
func (m Model) renderDiffPanel() string {
	files := m.filteredFiles()
	if len(files) == 0 {
		return RenderSplash(m.diffWidth(), m.diffHeight())
	}
	idx := m.opened
	if idx < 0 || idx >= len(files) {
		idx = 0
	}
	opts := RenderOpts{
		Width:          m.diffWidth(),
		ShowWhitespace: m.showWhitespace,
	}
	if m.searchActive {
		opts.Search = m.search
	}
	f := files[idx]
	body := m.renderFile(f, opts)
	return scrollLines(body, m.diffScroll, m.diffHeight())
}

func (m Model) renderFile(f diff.File, opts RenderOpts) string {
	if m.view == viewSideBySide {
		return RenderSideBySide(f, opts)
	}
	return RenderInline(f, opts)
}

func (m Model) diffWidth() int {
	// Both panels carry a 2-column border, so the body consumes
	// sidebarWidth + diffWidth + 4 horizontal cells.
	w := m.width - sidebarWidth - 4
	if w < 20 {
		w = 20
	}
	return w
}

// chromeRows counts the non-body lines: title bar, control bar, footer, and the
// optional search line.
func (m Model) chromeRows() int {
	rows := 3
	if m.searchActive {
		rows++
	}
	return rows
}

// diffHeight is the inner content height shared by both bordered panels. The
// border adds the two surrounding rows on top of this.
func (m Model) diffHeight() int {
	h := m.height - m.chromeRows() - 2 // 2 for the panel border's top/bottom
	if h < 1 {
		h = 1
	}
	return h
}

// padLinesTo trims s to n lines, padding with blank lines so the bordered panel
// always frames exactly n rows and both panels stay the same height.
func padLinesTo(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// scrollLines returns at most height lines of s starting at offset.
func scrollLines(s string, offset, height int) string {
	lines := strings.Split(s, "\n")
	if offset < 0 {
		offset = 0
	}
	if offset > len(lines) {
		offset = len(lines)
	}
	lines = lines[offset:]
	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
