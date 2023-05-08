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
		tag := newTag(test.input)
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
