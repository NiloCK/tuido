package tui

import (
	"testing"

	"github.com/nilock/tuido/tuido"
)

func focusTestItems() []*tuido.Item {
	mk := func(file string, line int, raw string) *tuido.Item {
		it := tuido.New(file, line, raw)
		return &it
	}
	return []*tuido.Item{
		mk("TODO.md", 1, "[@] Ship the parser rewrite ##file"),
		mk("TODO.md", 2, "[x] sketch the grammar"),
		mk("TODO.md", 3, "[ ] write the lexer"),
		mk("TODO.md", 4, "[ ] write the parser"),
		mk("notes.md", 1, "[ ] standalone item"),
	}
}

// In the aggregate view, a file with a ##file control item collapses to just
// that control item.
func TestRenderSelectionCollapsesControlledFile(t *testing.T) {
	tu := newTUI(focusTestItems(), runConfig)
	tu.populateRenderSelection()

	controlSeen, childSeen := false, false
	for _, it := range tu.renderSelection {
		if it.File() == "TODO.md" {
			if it.IsControl() {
				controlSeen = true
			} else {
				childSeen = true
			}
		}
	}

	if !controlSeen {
		t.Errorf("expected the TODO.md control item in aggregate view")
	}
	if childSeen {
		t.Errorf("expected TODO.md children to be collapsed in aggregate view")
	}
}

// Focus mode scopes the list to a single file and shows every item in it
// (including children and done items), in file order.
func TestFocusShowsAllFileItems(t *testing.T) {
	tu := newTUI(focusTestItems(), runConfig)
	tu.enterFocus("TODO.md")

	if len(tu.renderSelection) != 4 {
		t.Fatalf("expected 4 items in focus on TODO.md, got %d", len(tu.renderSelection))
	}
	for i, it := range tu.renderSelection {
		if it.File() != "TODO.md" {
			t.Errorf("focus leaked a non-TODO.md item: %q", it.File())
		}
		if i > 0 && tu.renderSelection[i-1].Line() > it.Line() {
			t.Errorf("focus items not in file order at index %d", i)
		}
	}

	// the checked child must be present (focus bypasses status filtering)
	foundDone := false
	for _, it := range tu.renderSelection {
		if it.Satus() == tuido.Checked {
			foundDone = true
		}
	}
	if !foundDone {
		t.Errorf("expected the checked child to appear in focus mode")
	}
}

// Exiting focus restores the collapsed aggregate view.
func TestExitFocusRestoresAggregate(t *testing.T) {
	tu := newTUI(focusTestItems(), runConfig)
	tu.enterFocus("TODO.md")
	tu.exitFocus()

	if tu.focused != "" {
		t.Errorf("expected focused to be cleared after exitFocus")
	}
	for _, it := range tu.renderSelection {
		if it.File() == "TODO.md" && !it.IsControl() {
			t.Errorf("expected children re-collapsed after exiting focus")
		}
	}
}
