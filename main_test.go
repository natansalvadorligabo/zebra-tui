package main

import (
	"path/filepath"
	"testing"
)

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
