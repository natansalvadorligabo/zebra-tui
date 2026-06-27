// Command zebra is a terminal UI for reviewing local git diffs.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/natansalvadorligabo/zebra-tui/internal/git"
	"github.com/natansalvadorligabo/zebra-tui/internal/ui"
)

// version is the build version, overridden at release time via
// -ldflags "-X main.version=...". It defaults to "dev" for local builds.
var version = "dev"

// writeVersion prints the program version to w.
func writeVersion(w io.Writer) {
	fmt.Fprintf(w, "zebra %s\n", version)
}

func main() {
	repoFlag := flag.String("repo", ".", "path to the git repository")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		writeVersion(os.Stdout)
		return
	}

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
	_, err = tea.NewProgram(model).Run()
	return err
}
