// Package diff parses unified git diff output into a structured model.
// It is pure: it performs no I/O and shells out to nothing.
package diff

import (
	"strconv"
	"strings"
)

// Status describes how a file changed relative to the base.
type Status int

const (
	StatusModified Status = iota
	StatusAdded
	StatusDeleted
	StatusRenamed
)

func (s Status) String() string {
	switch s {
	case StatusAdded:
		return "A"
	case StatusDeleted:
		return "D"
	case StatusRenamed:
		return "R"
	default:
		return "M"
	}
}

// LineType is the role of a line within a hunk.
type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
)

// Line is a single line within a hunk.
type Line struct {
	Type    LineType
	Content string // text without the leading +/-/space marker
	// OldNumber is the 1-based line number in the original file, or 0 if the
	// line does not exist there (i.e. an added line).
	OldNumber int
	// NewNumber is the 1-based line number in the new file, or 0 if the line
	// does not exist there (i.e. a removed line).
	NewNumber int
	// WhitespaceOnly is true when this added/removed line pairs with its
	// counterpart such that the only difference is whitespace.
	WhitespaceOnly bool
}

// Hunk is a contiguous group of changed lines.
type Hunk struct {
	Header   string // the raw "@@ -a,b +c,d @@" line
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []Line
}

// File is a single file's worth of changes.
type File struct {
	Path             string
	OldPath          string // populated for renames; otherwise equals Path
	Status           Status
	Added            int
	Removed          int
	IsBinary         bool
	BinarySizeBefore int64
	BinarySizeAfter  int64
	Hunks            []Hunk
}

// Parse transforms raw unified diff text into a slice of File models.
func Parse(raw string) ([]File, error) {
	var files []File
	lines := strings.Split(raw, "\n")

	var cur *File
	var hunk *Hunk
	var oldNo, newNo int

	flushHunk := func() {
		if cur != nil && hunk != nil {
			detectWhitespaceOnly(hunk)
			cur.Hunks = append(cur.Hunks, *hunk)
			hunk = nil
		}
	}
	flushFile := func() {
		flushHunk()
		if cur != nil {
			files = append(files, *cur)
			cur = nil
		}
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flushFile()
			f := File{Status: StatusModified}
			f.Path, f.OldPath = parseDiffGitPaths(line)
			cur = &f

		case cur == nil:
			// Preamble before the first file header; ignore.
			continue

		case strings.HasPrefix(line, "new file mode"):
			cur.Status = StatusAdded
		case strings.HasPrefix(line, "deleted file mode"):
			cur.Status = StatusDeleted
		case strings.HasPrefix(line, "rename from "):
			cur.Status = StatusRenamed
			cur.OldPath = strings.TrimPrefix(line, "rename from ")
		case strings.HasPrefix(line, "rename to "):
			cur.Status = StatusRenamed
			cur.Path = strings.TrimPrefix(line, "rename to ")

		case strings.HasPrefix(line, "Binary files ") && strings.HasSuffix(line, " differ"):
			cur.IsBinary = true

		case strings.HasPrefix(line, "--- "):
			// old-file marker; path already captured from the diff header.
		case strings.HasPrefix(line, "+++ "):
			// new-file marker.

		case strings.HasPrefix(line, "@@"):
			flushHunk()
			h := Hunk{Header: line}
			h.OldStart, h.OldLines, h.NewStart, h.NewLines = parseHunkHeader(line)
			oldNo, newNo = h.OldStart, h.NewStart
			hunk = &h

		case hunk != nil && strings.HasPrefix(line, "+"):
			hunk.Lines = append(hunk.Lines, Line{Type: LineAdded, Content: line[1:], NewNumber: newNo})
			newNo++
			cur.Added++
		case hunk != nil && strings.HasPrefix(line, "-"):
			hunk.Lines = append(hunk.Lines, Line{Type: LineRemoved, Content: line[1:], OldNumber: oldNo})
			oldNo++
			cur.Removed++
		case hunk != nil && strings.HasPrefix(line, " "):
			hunk.Lines = append(hunk.Lines, Line{Type: LineContext, Content: line[1:], OldNumber: oldNo, NewNumber: newNo})
			oldNo++
			newNo++
		}
	}
	flushFile()
	return files, nil
}

// detectWhitespaceOnly marks paired removed/added lines whose only difference
// is whitespace. It pairs each run of consecutive removed lines with the run of
// added lines that immediately follows, index by index.
func detectWhitespaceOnly(h *Hunk) {
	lines := h.Lines
	i := 0
	for i < len(lines) {
		if lines[i].Type != LineRemoved {
			i++
			continue
		}
		// Collect the run of removed lines.
		remStart := i
		for i < len(lines) && lines[i].Type == LineRemoved {
			i++
		}
		// Collect the run of added lines that immediately follows.
		addStart := i
		for i < len(lines) && lines[i].Type == LineAdded {
			i++
		}
		remCount := addStart - remStart
		addCount := i - addStart
		n := remCount
		if addCount < n {
			n = addCount
		}
		for k := 0; k < n; k++ {
			rem := &lines[remStart+k]
			add := &lines[addStart+k]
			if stripWhitespace(rem.Content) == stripWhitespace(add.Content) {
				rem.WhitespaceOnly = true
				add.WhitespaceOnly = true
			}
		}
	}
}

// stripWhitespace removes all whitespace characters from s.
func stripWhitespace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

// parseDiffGitPaths extracts the new and old path from a "diff --git a/x b/y" line.
func parseDiffGitPaths(line string) (newPath, oldPath string) {
	rest := strings.TrimPrefix(line, "diff --git ")
	fields := strings.SplitN(rest, " ", 2)
	if len(fields) != 2 {
		return "", ""
	}
	oldPath = strings.TrimPrefix(fields[0], "a/")
	newPath = strings.TrimPrefix(fields[1], "b/")
	return newPath, oldPath
}

// parseHunkHeader parses "@@ -oldStart,oldLines +newStart,newLines @@ ...".
func parseHunkHeader(line string) (oldStart, oldLines, newStart, newLines int) {
	// Strip the leading "@@ " and everything after the closing " @@".
	body := line
	if i := strings.Index(body, "@@"); i >= 0 {
		body = body[i+2:]
	}
	if i := strings.Index(body, "@@"); i >= 0 {
		body = body[:i]
	}
	body = strings.TrimSpace(body)
	for _, field := range strings.Fields(body) {
		switch {
		case strings.HasPrefix(field, "-"):
			oldStart, oldLines = parseRange(field[1:])
		case strings.HasPrefix(field, "+"):
			newStart, newLines = parseRange(field[1:])
		}
	}
	return
}

// parseRange parses "start,count" or just "start" (count defaults to 1).
func parseRange(s string) (start, count int) {
	count = 1
	parts := strings.SplitN(s, ",", 2)
	start, _ = strconv.Atoi(parts[0])
	if len(parts) == 2 {
		count, _ = strconv.Atoi(parts[1])
	}
	return
}
