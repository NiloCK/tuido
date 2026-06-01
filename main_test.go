package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles tuido into a temp dir and returns the binary path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "tuido")
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

// TestListCollapsesFileScopedItems asserts that `tuido list` collapses a file
// containing a ##file control item down to that control item (plus a child
// rollup), and does NOT emit the file-scoped siblings individually.
func TestListCollapsesFileScopedItems(t *testing.T) {
	bin := buildBinary(t)

	work := t.TempDir()
	content := strings.Join([]string{
		"[@] Ship the parser rewrite ##file",
		"[x] sketch the grammar",
		"[ ] write the lexer",
		"[ ] write the parser",
		"[ ] wire up errors",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(work, "TODO.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(bin, "list", work).CombinedOutput()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, out)
	}
	got := string(out)

	// the control item's text is shown...
	if !strings.Contains(got, "Ship the parser rewrite") {
		t.Errorf("expected control item in output, got:\n%s", got)
	}
	// ...with a child rollup (3 open/ongoing of 4 children)
	if !strings.Contains(got, "3 of 4") {
		t.Errorf("expected child rollup '3 of 4' in output, got:\n%s", got)
	}
	// but the file-scoped children must NOT be listed individually
	for _, child := range []string{"write the lexer", "write the parser", "wire up errors"} {
		if strings.Contains(got, child) {
			t.Errorf("child %q should be collapsed behind the control item, got:\n%s", child, got)
		}
	}
}

// TestListUncontrolledFileUnchanged asserts that files WITHOUT a control item
// still list every item, ie collapse is opt-in.
func TestListUncontrolledFileUnchanged(t *testing.T) {
	bin := buildBinary(t)

	work := t.TempDir()
	content := strings.Join([]string{
		"[ ] buy milk",
		"[ ] call the dentist",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(work, "list.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(bin, "list", work).CombinedOutput()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, out)
	}
	got := string(out)

	for _, item := range []string{"buy milk", "call the dentist"} {
		if !strings.Contains(got, item) {
			t.Errorf("expected %q in output of uncontrolled file, got:\n%s", item, got)
		}
	}
}
