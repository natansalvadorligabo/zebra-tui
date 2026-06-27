package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
)

// FilterFiles returns the files whose path contains the (case-insensitive)
// filter substring. An empty filter returns all files.
func FilterFiles(files []diff.File, filter string) []diff.File {
	if filter == "" {
		return files
	}
	q := strings.ToLower(filter)
	var out []diff.File
	for _, f := range files {
		if strings.Contains(strings.ToLower(f.Path), q) {
			out = append(out, f)
		}
	}
	return out
}

// RenderSidebar renders the file list with status indicators and change counts.
// selected is an index into the filtered list.
func RenderSidebar(files []diff.File, selected int, filter string, focused bool) string {
	files = FilterFiles(files, filter)
	var b strings.Builder
	title := "Files"
	if focused {
		title = styleFocused.Render("Files")
	}
	b.WriteString(title)
	b.WriteString("\n")
	if filter != "" {
		b.WriteString(styleMessage.Render("/" + filter))
		b.WriteString("\n")
	}
	for i, f := range files {
		row := sidebarRow(f)
		if i == selected {
			row = styleSelected.Render(row)
		}
		b.WriteString(row)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func sidebarRow(f diff.File) string {
	status := statusBadge(f.Status)
	counts := ""
	if f.Added > 0 {
		counts += " " + styleStatusA.Render(fmt.Sprintf("+%d", f.Added))
	}
	if f.Removed > 0 {
		counts += " " + styleStatusD.Render(fmt.Sprintf("-%d", f.Removed))
	}
	return fmt.Sprintf("%s %s%s", status, f.Path, counts)
}

func statusBadge(s diff.Status) string {
	switch s {
	case diff.StatusAdded:
		return styleStatusA.Render("A")
	case diff.StatusDeleted:
		return styleStatusD.Render("D")
	case diff.StatusRenamed:
		return styleStatusR.Render("R")
	default:
		return styleStatusM.Render("M")
	}
}

// RenderControlBar renders the persistent top bar showing current scope, view
// mode, and whitespace toggle. The field matching focus is highlighted so the
// active Tab target is obvious.
func RenderControlBar(scope, viewMode string, showWhitespace bool, focus focusPanel) string {
	ws := "off"
	if showWhitespace {
		ws = "on"
	}
	sep := styleMessage.Render(" │ ")
	field := func(label, value string, f focusPanel) string {
		if focus == f {
			return styleControlActive.Render(" ") +
				hotkeyLabel(styleControlActive, label) +
				styleControlActive.Render(": "+value+" ")
		}
		return " " + hotkeyLabel(styleMessage, label) + styleMessage.Render(": ") + styleFocused.Render(value) + " "
	}
	return field("scope", scope, focusScope) + sep + field("view", viewMode, focusView) + sep + field("whitespace", ws, focusWhitespace)
}

// hotkeyLabel renders label with base, underlining its first rune to advertise
// the single-key shortcut that toggles it (e.g. "scope" -> s̲cope, bound to "s").
func hotkeyLabel(base lipgloss.Style, label string) string {
	r := []rune(label)
	if len(r) == 0 {
		return ""
	}
	return base.Underline(true).Render(string(r[0])) + base.Render(string(r[1:]))
}

// RenderFooter renders the keybinding hint line.
func RenderFooter() string {
	return styleMessage.Render(
		"↹ focus  ←/→ panel  j/k scroll  n/p hunk  enter open  ^f search  q quit",
	)
}

// RenderEmptyMessage is shown in the diff panel when there is nothing to diff.
func RenderEmptyMessage() string {
	return styleMessage.Render(emptyMessage)
}
