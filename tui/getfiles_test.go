package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func containsPath(paths []string, target string) bool {
	for _, p := range paths {
		if p == target {
			return true
		}
	}
	return false
}

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempFile %s: %v", path, err)
	}
	return path
}

func TestGetFiles_readsItemWithNoConfig(t *testing.T) {
	dir := t.TempDir()
	want := writeTempFile(t, dir, "ignoreme.xit", "[ ] a task\n")

	got := getFiles(dir, []string{"xit"}, nil)

	if !containsPath(got, want) {
		t.Errorf("expected %s in results, got %v", want, got)
	}
}

func TestGetFiles_excludeFile(t *testing.T) {
	dir := t.TempDir()
	excluded := writeTempFile(t, dir, "ignoreme.xit", "[ ] a task\n")
	kept := writeTempFile(t, dir, "keepme.xit", "[ ] another task\n")

	got := getFiles(dir, []string{"xit"}, []string{"ignoreme.xit"})

	if containsPath(got, excluded) {
		t.Errorf("expected %s to be excluded, got %v", excluded, got)
	}
	if !containsPath(got, kept) {
		t.Errorf("expected %s to be kept, got %v", kept, got)
	}
}

func TestGetFiles_excludeDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "ignorethisdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	excluded := writeTempFile(t, subdir, "item.xit", "[ ] excluded task\n")
	kept := writeTempFile(t, dir, "keepme.xit", "[ ] kept task\n")

	got := getFiles(dir, []string{"xit"}, []string{"ignorethisdir"})

	if containsPath(got, excluded) {
		t.Errorf("expected %s to be excluded, got %v", excluded, got)
	}
	if !containsPath(got, kept) {
		t.Errorf("expected %s to be kept, got %v", kept, got)
	}
}

func TestGetFiles_excludeGlob(t *testing.T) {
	dir := t.TempDir()
	excluded1 := writeTempFile(t, dir, "ignore_this.xit", "[ ] task one\n")
	excluded2 := writeTempFile(t, dir, "ignore_that.xit", "[ ] task two\n")
	kept := writeTempFile(t, dir, "keepme.xit", "[ ] kept task\n")

	got := getFiles(dir, []string{"xit"}, []string{"ignore_*"})

	for _, ex := range []string{excluded1, excluded2} {
		if containsPath(got, ex) {
			t.Errorf("expected %s to be excluded by glob, got %v", ex, got)
		}
	}
	if !containsPath(got, kept) {
		t.Errorf("expected %s to be kept, got %v", kept, got)
	}
}

func TestGetFiles_excludeViaSubdirConfig(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	ignoredDir := filepath.Join(subdir, "ignorethisdir")
	if err := os.Mkdir(ignoredDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeTempFile(t, subdir, ".tuido", "exclude=ignorethisdir\n")
	excluded := writeTempFile(t, ignoredDir, "item.xit", "[ ] excluded task\n")
	kept := writeTempFile(t, subdir, "keepme.xit", "[ ] kept task\n")

	got := getFiles(dir, []string{"xit"}, nil)

	if containsPath(got, excluded) {
		t.Errorf("expected %s to be excluded via .tuido config, got %v", excluded, got)
	}
	if !containsPath(got, kept) {
		t.Errorf("expected %s to be kept, got %v", kept, got)
	}
}
