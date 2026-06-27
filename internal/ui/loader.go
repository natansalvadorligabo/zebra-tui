package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
	"github.com/salvadorligabo/zebra-tui/internal/git"
)

// filesLoadedMsg carries the parsed diff for a scope back into the model.
type filesLoadedMsg struct {
	files []diff.File
	err   error
}

// LoadFiles runs git for the given scope, parses the diff, and enriches binary
// files with their before/after sizes. It is the single I/O seam that ties the
// pure parser to the git wrapper.
func LoadFiles(repo string, scope git.Scope) ([]diff.File, error) {
	raw, err := git.Diff(repo, scope)
	if err != nil {
		return nil, err
	}
	files, err := diff.Parse(raw)
	if err != nil {
		return nil, err
	}
	for i := range files {
		if files[i].IsBinary {
			before, after, _ := git.FileSizes(repo, files[i].Path)
			files[i].BinarySizeBefore = before
			files[i].BinarySizeAfter = after
		}
	}
	return files, nil
}

// loadFilesCmd is the tea.Cmd form of LoadFiles.
func loadFilesCmd(repo string, scope git.Scope) tea.Cmd {
	return func() tea.Msg {
		files, err := LoadFiles(repo, scope)
		return filesLoadedMsg{files: files, err: err}
	}
}
