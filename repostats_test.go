package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCountLines(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		content string
		want    int
	}{
		{"", 0},
		{"one line\n", 1},
		{"a\nb\nc\n", 3},
		{"no trailing newline", 0},
		{"a\nb", 1},
	}
	for i, tt := range tests {
		path := filepath.Join(dir, fmt.Sprintf("file%d", i))
		if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
		got, err := countLines(path)
		if err != nil {
			t.Fatalf("countLines(%q): %v", tt.content, err)
		}
		if got != tt.want {
			t.Errorf("countLines(%q) = %d, want %d", tt.content, got, tt.want)
		}
	}
}

func TestCollectRepoStats(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatalf("writing %s: %v", rel, err)
		}
	}
	// Two regular files with 2 and 1 lines respectively.
	write("a.txt", "1\n2\n")
	write("sub/b.txt", "x\n")
	// Files under .git must be skipped entirely.
	write(".git/config", "junk\n")
	write(".git/refs/head", "z\n")

	s, err := collectRepoStats(dir)
	if err != nil {
		t.Fatalf("collectRepoStats: %v", err)
	}
	if s.fileCount != 2 {
		t.Errorf("fileCount = %d, want 2 (.git must be skipped)", s.fileCount)
	}
	if s.totalLines != 3 {
		t.Errorf("totalLines = %d, want 3", s.totalLines)
	}
}

func TestLastCommitDateNonRepo(t *testing.T) {
	// A fresh temp directory is not a git repository, so the date is unknown.
	if got := lastCommitDate(t.TempDir()); got != "unknown" {
		t.Errorf("lastCommitDate(non-repo) = %q, want %q", got, "unknown")
	}
}

func TestWriteRepoStats(t *testing.T) {
	var buf bytes.Buffer
	writeRepoStats(&buf, repoStats{fileCount: 5, totalLines: 100, lastCommit: "2026-07-25"})
	want := "files:       5\nlines:       100\nlast commit: 2026-07-25\n"
	if got := buf.String(); got != want {
		t.Errorf("writeRepoStats output =\n%q\nwant\n%q", got, want)
	}
}
