package diff

import "testing"

func TestParse_SimpleModification(t *testing.T) {
	raw := `diff --git a/hello.go b/hello.go
index 1234567..89abcde 100644
--- a/hello.go
+++ b/hello.go
@@ -1,3 +1,3 @@
 package main
-func old() {}
+func new() {}
 // trailing
`
	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.Path != "hello.go" {
		t.Errorf("Path = %q, want %q", f.Path, "hello.go")
	}
	if f.Status != StatusModified {
		t.Errorf("Status = %v, want Modified", f.Status)
	}
	if f.Added != 1 || f.Removed != 1 {
		t.Errorf("Added/Removed = %d/%d, want 1/1", f.Added, f.Removed)
	}
	if len(f.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
	}
	h := f.Hunks[0]
	wantTypes := []LineType{LineContext, LineRemoved, LineAdded, LineContext}
	if len(h.Lines) != len(wantTypes) {
		t.Fatalf("expected %d lines, got %d", len(wantTypes), len(h.Lines))
	}
	for i, wt := range wantTypes {
		if h.Lines[i].Type != wt {
			t.Errorf("line %d type = %v, want %v", i, h.Lines[i].Type, wt)
		}
	}
	// Line content excludes the +/-/space prefix.
	if h.Lines[1].Content != "func old() {}" {
		t.Errorf("removed content = %q", h.Lines[1].Content)
	}
	if h.Lines[2].Content != "func new() {}" {
		t.Errorf("added content = %q", h.Lines[2].Content)
	}
	// Line numbering: context line "package main" is old 1 / new 1.
	if h.Lines[0].OldNumber != 1 || h.Lines[0].NewNumber != 1 {
		t.Errorf("context line numbers = %d/%d, want 1/1", h.Lines[0].OldNumber, h.Lines[0].NewNumber)
	}
	// removed line has old number 2, new number 0
	if h.Lines[1].OldNumber != 2 || h.Lines[1].NewNumber != 0 {
		t.Errorf("removed line numbers = %d/%d, want 2/0", h.Lines[1].OldNumber, h.Lines[1].NewNumber)
	}
	// added line has old number 0, new number 2
	if h.Lines[2].OldNumber != 0 || h.Lines[2].NewNumber != 2 {
		t.Errorf("added line numbers = %d/%d, want 0/2", h.Lines[2].OldNumber, h.Lines[2].NewNumber)
	}
}

func TestParse_Empty(t *testing.T) {
	files, err := Parse("")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestParse_AddedFile(t *testing.T) {
	raw := `diff --git a/new.txt b/new.txt
new file mode 100644
index 0000000..3b18e51
--- /dev/null
+++ b/new.txt
@@ -0,0 +1,2 @@
+hello
+world
`
	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.Status != StatusAdded {
		t.Errorf("Status = %v, want Added", f.Status)
	}
	if f.Added != 2 || f.Removed != 0 {
		t.Errorf("Added/Removed = %d/%d, want 2/0", f.Added, f.Removed)
	}
	if f.Hunks[0].Lines[0].NewNumber != 1 || f.Hunks[0].Lines[1].NewNumber != 2 {
		t.Errorf("new numbers = %d,%d want 1,2", f.Hunks[0].Lines[0].NewNumber, f.Hunks[0].Lines[1].NewNumber)
	}
}

func TestParse_DeletedFile(t *testing.T) {
	raw := `diff --git a/gone.txt b/gone.txt
deleted file mode 100644
index 3b18e51..0000000
--- a/gone.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-hello
-world
`
	files, _ := Parse(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Status != StatusDeleted {
		t.Errorf("Status = %v, want Deleted", files[0].Status)
	}
	if files[0].Removed != 2 {
		t.Errorf("Removed = %d, want 2", files[0].Removed)
	}
}

func TestParse_RenamedFile(t *testing.T) {
	raw := `diff --git a/old/name.txt b/new/name.txt
similarity index 92%
rename from old/name.txt
rename to new/name.txt
index 1111111..2222222 100644
--- a/old/name.txt
+++ b/new/name.txt
@@ -1,1 +1,1 @@
-old line
+new line
`
	files, _ := Parse(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.Status != StatusRenamed {
		t.Errorf("Status = %v, want Renamed", f.Status)
	}
	if f.OldPath != "old/name.txt" {
		t.Errorf("OldPath = %q, want old/name.txt", f.OldPath)
	}
	if f.Path != "new/name.txt" {
		t.Errorf("Path = %q, want new/name.txt", f.Path)
	}
}

func TestParse_BinaryFile(t *testing.T) {
	raw := `diff --git a/logo.png b/logo.png
index 1111111..2222222 100644
Binary files a/logo.png and b/logo.png differ
`
	files, _ := Parse(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsBinary {
		t.Errorf("IsBinary = false, want true")
	}
	if len(files[0].Hunks) != 0 {
		t.Errorf("binary file should have no hunks, got %d", len(files[0].Hunks))
	}
}

func TestParse_MultiHunk(t *testing.T) {
	raw := `diff --git a/multi.txt b/multi.txt
index 1111111..2222222 100644
--- a/multi.txt
+++ b/multi.txt
@@ -1,3 +1,3 @@
 a
-b
+B
 c
@@ -10,3 +10,3 @@
 x
-y
+Y
 z
`
	files, _ := Parse(raw)
	if len(files[0].Hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(files[0].Hunks))
	}
	if files[0].Hunks[1].OldStart != 10 || files[0].Hunks[1].NewStart != 10 {
		t.Errorf("second hunk starts = %d/%d, want 10/10", files[0].Hunks[1].OldStart, files[0].Hunks[1].NewStart)
	}
	if files[0].Added != 2 || files[0].Removed != 2 {
		t.Errorf("Added/Removed = %d/%d, want 2/2", files[0].Added, files[0].Removed)
	}
}

func TestParse_WhitespaceOnlyChange(t *testing.T) {
	// Only indentation changes: tabs added before the two middle lines.
	raw := "diff --git a/ws.go b/ws.go\n" +
		"index 1111111..2222222 100644\n" +
		"--- a/ws.go\n" +
		"+++ b/ws.go\n" +
		"@@ -1,4 +1,4 @@\n" +
		" func f() {\n" +
		"-x := 1\n" +
		"-y := 2\n" +
		"+\tx := 1\n" +
		"+\ty := 2\n" +
		" }\n"
	files, _ := Parse(raw)
	h := files[0].Hunks[0]
	// lines: [0]=context, [1]=removed x, [2]=removed y, [3]=added x, [4]=added y, [5]=context
	for _, idx := range []int{1, 2, 3, 4} {
		if !h.Lines[idx].WhitespaceOnly {
			t.Errorf("line %d (%q) WhitespaceOnly = false, want true", idx, h.Lines[idx].Content)
		}
	}
	if h.Lines[0].WhitespaceOnly {
		t.Errorf("context line should not be WhitespaceOnly")
	}
}

func TestParse_RealChangeNotWhitespaceOnly(t *testing.T) {
	files, _ := Parse(`diff --git a/r.go b/r.go
index 1111111..2222222 100644
--- a/r.go
+++ b/r.go
@@ -1,1 +1,1 @@
-x := 1
+x := 2
`)
	for _, l := range files[0].Hunks[0].Lines {
		if l.WhitespaceOnly {
			t.Errorf("line %q should not be WhitespaceOnly", l.Content)
		}
	}
}
