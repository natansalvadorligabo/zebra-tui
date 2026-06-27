package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteVersion_DefaultsToDev(t *testing.T) {
	var buf bytes.Buffer
	writeVersion(&buf)
	got := buf.String()
	if !strings.HasPrefix(got, "zebra ") {
		t.Errorf("writeVersion output = %q, want it to start with %q", got, "zebra ")
	}
	if !strings.Contains(got, "dev") {
		t.Errorf("writeVersion output = %q, want it to contain default version %q", got, "dev")
	}
}

func TestResolveRepo_Default(t *testing.T) {
	got, err := resolveRepo(".")
	if err != nil {
		t.Fatalf("resolveRepo error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("resolveRepo(%q) = %q, want absolute path", ".", got)
	}
}

func TestResolveRepo_RelativeBecomesAbsolute(t *testing.T) {
	got, err := resolveRepo("some/sub")
	if err != nil {
		t.Fatalf("resolveRepo error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("resolveRepo result = %q, want absolute", got)
	}
	if filepath.Base(got) != "sub" {
		t.Errorf("resolveRepo result = %q, want path ending in 'sub'", got)
	}
}
