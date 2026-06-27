package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newRepo creates a fresh git repo in a temp dir with one committed file.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("config", "commit.gpgsign", "false")
	writeFile(t, dir, "file.txt", "line1\nline2\nline3\n")
	run("add", ".")
	run("commit", "-m", "initial")
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitAdd(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

func TestDiff_WorkingTree(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "file.txt", "line1\nCHANGED\nline3\n")

	out, err := Diff(dir, ScopeWorkingTree)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if !strings.Contains(out, "diff --git a/file.txt b/file.txt") {
		t.Errorf("missing diff header in output:\n%s", out)
	}
	if !strings.Contains(out, "+CHANGED") || !strings.Contains(out, "-line2") {
		t.Errorf("missing change lines in output:\n%s", out)
	}
}

func TestDiff_WorkingTreeIgnoresStaged(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "file.txt", "line1\nSTAGED\nline3\n")
	gitAdd(t, dir)

	out, err := Diff(dir, ScopeWorkingTree)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("working-tree diff should be empty when all changes are staged, got:\n%s", out)
	}
}

func TestDiff_Staged(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "file.txt", "line1\nSTAGED\nline3\n")
	gitAdd(t, dir)

	out, err := Diff(dir, ScopeStaged)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if !strings.Contains(out, "+STAGED") {
		t.Errorf("staged diff missing +STAGED:\n%s", out)
	}
}

func TestDiff_All(t *testing.T) {
	dir := newRepo(t)
	// Commit a second tracked file so we can leave an unstaged change on it.
	writeFile(t, dir, "other.txt", "orig\n")
	gitAdd(t, dir)
	gitCommit(t, dir, "add other")

	// One staged change and one unstaged change to different tracked files.
	writeFile(t, dir, "file.txt", "line1\nSTAGED\nline3\n")
	gitAdd(t, dir)
	writeFile(t, dir, "other.txt", "UNSTAGED\n")

	out, err := Diff(dir, ScopeAll)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if !strings.Contains(out, "+STAGED") {
		t.Errorf("All diff missing staged change:\n%s", out)
	}
	if !strings.Contains(out, "+UNSTAGED") {
		t.Errorf("All diff missing unstaged change:\n%s", out)
	}
}

func TestDiff_CleanWorkingTree(t *testing.T) {
	dir := newRepo(t)
	out, err := Diff(dir, ScopeWorkingTree)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("clean working tree should produce empty diff, got:\n%s", out)
	}
}

func TestDiff_NotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := Diff(dir, ScopeWorkingTree)
	if err == nil {
		t.Errorf("expected error for non-git directory, got nil")
	}
}

func TestFileSizes(t *testing.T) {
	dir := newRepo(t)
	// Commit a "binary" blob of known size, then grow it in the working tree.
	before := strings.Repeat("A", 100)
	writeFile(t, dir, "blob.bin", before)
	gitAdd(t, dir)
	gitCommit(t, dir, "add blob")
	writeFile(t, dir, "blob.bin", strings.Repeat("A", 250))

	b, a, err := FileSizes(dir, "blob.bin")
	if err != nil {
		t.Fatalf("FileSizes error: %v", err)
	}
	if b != 100 {
		t.Errorf("before size = %d, want 100", b)
	}
	if a != 250 {
		t.Errorf("after size = %d, want 250", a)
	}
}

func TestIsRepo(t *testing.T) {
	if !IsRepo(newRepo(t)) {
		t.Errorf("IsRepo on a git repo = false, want true")
	}
	if IsRepo(t.TempDir()) {
		t.Errorf("IsRepo on a plain dir = true, want false")
	}
}
