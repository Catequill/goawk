// repostats.go implements the -repostats flag, which prints a short
// summary of the repository in the current working directory: the number
// of regular files, the total number of lines across those files, and the
// date of the most recent git commit.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// repoStats holds summary statistics about a repository directory tree:
// the number of regular files, the total number of lines across those
// files, and the short date of the most recent git commit.
type repoStats struct {
	fileCount  int
	totalLines int
	lastCommit string
}

// collectRepoStats walks the directory tree rooted at root, counting
// regular files and their combined line count (newline bytes, matching
// "wc -l"). The .git directory is skipped so version-control metadata is
// not included. The last commit date is read from git via lastCommitDate.
func collectRepoStats(root string) (repoStats, error) {
	var s repoStats
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		s.fileCount++
		lines, err := countLines(path)
		if err != nil {
			return err
		}
		s.totalLines += lines
		return nil
	})
	if err != nil {
		return repoStats{}, err
	}
	s.lastCommit = lastCommitDate(root)
	return s, nil
}

// countLines returns the number of newline bytes in the file at path,
// matching the line count reported by "wc -l".
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var count int
	buf := make([]byte, 64*1024)
	for {
		n, err := f.Read(buf)
		count += bytes.Count(buf[:n], []byte{'\n'})
		if err == io.EOF {
			return count, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

// lastCommitDate returns the short date (YYYY-MM-DD) of the most recent
// git commit in the repository rooted at root. If git is unavailable or
// root is not a git repository, it returns "unknown".
func lastCommitDate(root string) string {
	cmd := exec.Command("git", "-C", root, "log", "-1", "--format=%cd", "--date=short")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	date := strings.TrimSpace(string(out))
	if date == "" {
		return "unknown"
	}
	return date
}

// writeRepoStats formats the repository statistics as aligned key/value
// lines and writes them to w.
func writeRepoStats(w io.Writer, s repoStats) {
	fmt.Fprintf(w, "files:       %d\n", s.fileCount)
	fmt.Fprintf(w, "lines:       %d\n", s.totalLines)
	fmt.Fprintf(w, "last commit: %s\n", s.lastCommit)
}

// printRepoStats collects statistics for the current working directory and
// writes them to w, returning an error if the tree cannot be walked.
func printRepoStats(w io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	s, err := collectRepoStats(dir)
	if err != nil {
		return err
	}
	writeRepoStats(w, s)
	return nil
}
