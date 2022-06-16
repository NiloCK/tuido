package tuido

import (
	"testing"
)

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
