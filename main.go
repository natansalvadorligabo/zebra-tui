// Command zebra is a terminal UI for reviewing local git diffs.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/salvadorligabo/zebra-tui/internal/git"
	"github.com/salvadorligabo/zebra-tui/internal/ui"
)

func main() {
	repoFlag := flag.String("repo", ".", "path to the git repository")
	flag.Parse()

	if err := run(*repoFlag); err != nil {
		fmt.Fprintln(os.Stderr, "zebra:", err)
		os.Exit(1)
	}
}

// resolveRepo turns a (possibly relative) repo flag into an absolute path.
func resolveRepo(path string) (string, error) {
	return filepath.Abs(path)
}

func run(repoFlag string) error {
	repo, err := resolveRepo(repoFlag)
	if err != nil {
		return err
	}

	if !git.IsRepo(repo) {
		return fmt.Errorf("not a git repository: %s", repo)
	}

	// Load the working-tree diff up front so the first frame is instant.
	scope := git.ScopeWorkingTree
	files, err := ui.LoadFiles(repo, scope)
	if err != nil {
		return err
	}

	model := ui.NewWithFiles(repo, scope, files)
	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	return err
}
