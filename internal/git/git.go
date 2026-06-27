// Package git is a thin wrapper around the git binary. It runs diff commands
// and returns their raw output; it does not parse diffs.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Scope selects which set of changes a diff covers.
type Scope int

const (
	// ScopeWorkingTree is unstaged changes (git diff).
	ScopeWorkingTree Scope = iota
	// ScopeStaged is changes staged for commit (git diff --cached).
	ScopeStaged
	// ScopeAll is working tree plus staged changes against HEAD (git diff HEAD).
	ScopeAll
)

// String returns a short human-readable label for the scope.
func (s Scope) String() string {
	switch s {
	case ScopeStaged:
		return "staged"
	case ScopeAll:
		return "worktree+staged"
	default:
		return "worktree"
	}
}

// args returns the git arguments that produce the diff for this scope.
func (s Scope) args() []string {
	switch s {
	case ScopeStaged:
		return []string{"diff", "--cached"}
	case ScopeAll:
		return []string{"diff", "HEAD"}
	default:
		return []string{"diff"}
	}
}

// Diff returns the raw unified diff for the given repository and scope.
func Diff(repo string, scope Scope) (string, error) {
	return run(repo, scope.args()...)
}

// IsRepo reports whether path is inside a git working tree.
func IsRepo(path string) bool {
	out, err := run(path, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// FileSizes returns the byte size of path at HEAD (before) and in the working
// tree (after). A missing HEAD blob (newly added file) yields before = 0; a
// missing working-tree file (deleted) yields after = 0.
func FileSizes(repo, path string) (before, after int64, err error) {
	if out, e := run(repo, "cat-file", "-s", "HEAD:"+path); e == nil {
		before, _ = strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	}
	if fi, e := os.Stat(filepath.Join(repo, path)); e == nil {
		after = fi.Size()
	}
	return before, after, nil
}

// run executes git with the given args in repo and returns stdout.
func run(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
