package tuido

import (
	"testing"
)

func TestNewTag(t *testing.T) {
	type tc struct {
		input string
		name  string
		value string
		str   string
	}

	tests := []tc{
		{
			input: "foo1",
			name:  "foo1",
			value: "",
			str:   "foo1",
		},
		{
			input: "foo2=bar",
			name:  "foo2",
			value: "bar",
			str:   "foo2=bar",
		},
		{
			input: "foo3=bar=bar",
			name:  "foo3",
			value: "bar=bar",
			str:   "foo3=bar=bar",
		},
		{
			input: "#extrapound",
			name:  "extrapound",
			value: "",
			str:   "extrapound",
		},
	}

	for _, test := range tests {
		tag := NewTag(test.input)
		if tag.name != test.name {
			t.Errorf("expected tag name %s, but found %s", test.name, tag.name)
		}
		if tag.Name() != test.name {
			t.Errorf("expected tag string %s, but found %s", test.input, tag.String())
		}
		if tag.String() != test.str {
			t.Errorf("expected tag string %s, but found %s", test.input, tag.String())
		}
		if tag.value != test.value {
			t.Errorf("expected tag value %s, but found %s", test.value, tag.value)
		}
	}

}

func TestSystemTagsExcludedFromUserTags(t *testing.T) {
	i := Item{raw: "[ ] do the thing #real ##file ##scope=docs"}

	// ## tags must not leak into the user tag space
	for _, tag := range i.Tags() {
		if tag.Name() == "file" || tag.Name() == "scope" {
			t.Errorf("system tag %q leaked into user Tags()", tag.Name())
		}
	}

	// the genuine user tag is still present
	foundReal := false
	for _, tag := range i.Tags() {
		if tag.Name() == "real" {
			foundReal = true
		}
	}
	if !foundReal {
		t.Errorf("expected user tag 'real' in Tags(), got %v", i.Tags())
	}

	// system tags are parsed, with values
	sys := i.SystemTags()
	if len(sys) != 2 {
		t.Fatalf("expected 2 system tags, got %d (%v)", len(sys), sys)
	}
	if sys[0].Name() != "file" || sys[0].Value() != "" {
		t.Errorf("expected ##file with no value, got %q=%q", sys[0].Name(), sys[0].Value())
	}
	if sys[1].Name() != "scope" || sys[1].Value() != "docs" {
		t.Errorf("expected ##scope=docs, got %q=%q", sys[1].Name(), sys[1].Value())
	}
}

func TestIsControl(t *testing.T) {
	control := Item{raw: "[@] Ship the parser rewrite ##file"}
	plain := Item{raw: "[ ] write the lexer #due=2026-01-01"}

	if !control.IsControl() {
		t.Errorf("expected ##file item to be a control item")
	}
	if plain.IsControl() {
		t.Errorf("did not expect plain item to be a control item")
	}
}

func TestCollapseFileScoped(t *testing.T) {
	mk := func(file, raw string) *Item { return &Item{file: file, raw: raw} }
	items := []*Item{
		mk("TODO.md", "[@] Ship the parser rewrite ##file"),
		mk("TODO.md", "[x] sketch the grammar"),
		mk("TODO.md", "[ ] write the lexer"),
		mk("TODO.md", "[ ] write the parser"),
		mk("notes.md", "[ ] standalone item"),
	}

	collapsed := CollapseFileScoped(items)

	if len(collapsed) != 2 {
		t.Fatalf("expected 2 items after collapse (control + standalone), got %d", len(collapsed))
	}
	if !collapsed[0].IsControl() {
		t.Errorf("expected the control item to survive collapse")
	}
	if collapsed[1].Text() != "standalone item" {
		t.Errorf("expected uncontrolled file's item to survive, got %q", collapsed[1].Text())
	}

	// rollup: 2 of 3 children open/ongoing (grammar is checked)
	rem, tot := ChildStats(items[0], items)
	if rem != 2 || tot != 3 {
		t.Errorf("expected child stats (2, 3), got (%d, %d)", rem, tot)
	}
}

func TestImportance(t *testing.T) {
	items := []Item{
		{
			file: "",
			line: -1,
			raw:  "[ ] not important at all",
		},
		{
			"",
			-1,
			"[ ] ! a bit important",
		},
		{
			"", -1, "[ ] !! a little more",
		},
		{
			"", -1, "[ ] ..!!! has leading periods, but should still be 3",
		},
	}

	for i, item := range items {
		if item.Importance() != i {
			t.Errorf("expected importance %d, but found %d", i, item.Importance())
		}
	}
}
