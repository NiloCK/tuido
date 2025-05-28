package tui

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/nilock/tuido/tuido"
)

func TestParseTagFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected tuido.Tag
	}{
		{
			name:     "simple tag",
			input:    "#due",
			expected: tuido.NewTag("due"),
		},
		{
			name:     "tag with value",
			input:    "#due=2024-01-15",
			expected: tuido.NewTag("due=2024-01-15"),
		},
		{
			name:     "tag with complex value",
			input:    "#project=work-stuff",
			expected: tuido.NewTag("project=work-stuff"),
		},
		{
			name:     "tag with equals in value",
			input:    "#query=x=y",
			expected: tuido.NewTag("query=x=y"),
		},
		{
			name:     "single character tag",
			input:    "#a",
			expected: tuido.NewTag("a"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTagFilter(tt.input)
			if result.Name() != tt.expected.Name() || result.String() != tt.expected.String() {
				t.Errorf("parseTagFilter(%s) = %+v, want %+v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestItemHasTag(t *testing.T) {
	// Create test tags for items
	itemTags := []tuido.Tag{
		tuido.NewTag("due=2024-01-15"),
		tuido.NewTag("project=work"),
		tuido.NewTag("active"),
		tuido.NewTag("important"),
	}

	tests := []struct {
		name     string
		filter   tuido.Tag
		expected bool
	}{
		{
			name:     "exact tag and value match",
			filter:   tuido.NewTag("due=2024-01-15"),
			expected: true,
		},
		{
			name:     "tag name only match (any value)",
			filter:   tuido.NewTag("due"),
			expected: true,
		},
		{
			name:     "tag name match different value",
			filter:   tuido.NewTag("due=2024-01-16"),
			expected: false,
		},
		{
			name:     "tag without value matches tag without value",
			filter:   tuido.NewTag("active"),
			expected: true,
		},
		{
			name:     "tag name only matches tag without value",
			filter:   tuido.NewTag("important"),
			expected: true,
		},
		{
			name:     "no match - different tag name",
			filter:   tuido.NewTag("nonexistent"),
			expected: false,
		},
		{
			name:     "no match - different tag name with value",
			filter:   tuido.NewTag("nonexistent=value"),
			expected: false,
		},
		// New tests for starts-with matching
		{
			name:     "partial tag name match - single char",
			filter:   tuido.NewTag("d"),
			expected: true, // matches "due=2024-01-15"
		},
		{
			name:     "partial tag name match - multiple chars",
			filter:   tuido.NewTag("du"),
			expected: true, // matches "due=2024-01-15"
		},
		{
			name:     "partial tag name match - prefix of tag without value",
			filter:   tuido.NewTag("act"),
			expected: true, // matches "active"
		},
		{
			name:     "partial tag name match - prefix of project",
			filter:   tuido.NewTag("proj"),
			expected: true, // matches "project=work"
		},
		{
			name:     "partial tag name match - single char for important",
			filter:   tuido.NewTag("i"),
			expected: true, // matches "important"
		},
		{
			name:     "no partial match - wrong prefix",
			filter:   tuido.NewTag("xyz"),
			expected: false,
		},
		{
			name:     "no partial match - longer than existing tag",
			filter:   tuido.NewTag("activelyworking"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := itemHasTag(itemTags, tt.filter)
			if result != tt.expected {
				t.Errorf("itemHasTag(%+v, %+v) = %v, want %v", 
					itemTags, tt.filter, result, tt.expected)
			}
		})
	}
}

func TestItemMatchesAllTags(t *testing.T) {
	tests := []struct {
		name        string
		itemText    string
		tagFilters  []tuido.Tag
		expected    bool
	}{
		{
			name:        "single tag match",
			itemText:    "fix bug #due=2024-01-15",
			tagFilters:  []tuido.Tag{tuido.NewTag("due")},
			expected:    true,
		},
		{
			name:        "single tag with value match",
			itemText:    "fix bug #due=2024-01-15",
			tagFilters:  []tuido.Tag{tuido.NewTag("due=2024-01-15")},
			expected:    true,
		},
		{
			name:        "multiple tags all match",
			itemText:    "fix bug #due=2024-01-15 #project=work #active",
			tagFilters:  []tuido.Tag{tuido.NewTag("due"), tuido.NewTag("project=work")},
			expected:    true,
		},
		{
			name:        "one tag missing",
			itemText:    "fix bug #due=2024-01-15",
			tagFilters:  []tuido.Tag{tuido.NewTag("due"), tuido.NewTag("project")},
			expected:    false,
		},
		{
			name:        "wrong tag value",
			itemText:    "fix bug #due=2024-01-15",
			tagFilters:  []tuido.Tag{tuido.NewTag("due=2024-01-16")},
			expected:    false,
		},
		{
			name:        "no tag filters",
			itemText:    "fix bug #due=2024-01-15",
			tagFilters:  []tuido.Tag{},
			expected:    true,
		},
		// New tests for starts-with matching
		{
			name:        "partial tag match - single char",
			itemText:    "fix bug #due=2024-01-15 #project=work",
			tagFilters:  []tuido.Tag{tuido.NewTag("d")},
			expected:    true,
		},
		{
			name:        "partial tag match - multiple chars",
			itemText:    "fix bug #due=2024-01-15 #project=work",
			tagFilters:  []tuido.Tag{tuido.NewTag("proj")},
			expected:    true,
		},
		{
			name:        "multiple partial tag matches",
			itemText:    "fix bug #due=2024-01-15 #project=work #active",
			tagFilters:  []tuido.Tag{tuido.NewTag("d"), tuido.NewTag("act")},
			expected:    true,
		},
		{
			name:        "partial match fails when one doesn't match",
			itemText:    "fix bug #due=2024-01-15",
			tagFilters:  []tuido.Tag{tuido.NewTag("d"), tuido.NewTag("proj")},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary item with the test text
			testItem := tuido.New("test.txt", 1, "[ ] "+tt.itemText)
			
			result := itemMatchesAllTags(&testItem, tt.tagFilters)
			if result != tt.expected {
				t.Errorf("itemMatchesAllTags(item with text %q, %+v) = %v, want %v", 
					tt.itemText, tt.tagFilters, result, tt.expected)
			}
		})
	}
}

func TestApplyTagFilters(t *testing.T) {
	// Create test items
	items := []*tuido.Item{
		func() *tuido.Item { item := tuido.New("test.txt", 1, "[ ] fix bug #due=2024-01-15 #project=work"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 2, "[ ] write docs #due=2024-01-16 #project=docs"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 3, "[ ] review code #project=work #active"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 4, "[ ] meeting tomorrow #active"); return &item }(),
	}

	// Create a test TUI instance
	testTui := &tui{}

	tests := []struct {
		name           string
		tagFilters     []tuido.Tag
		expectedCount  int
		expectedItems  []int // indices of items that should match
	}{
		{
			name:           "no filters returns all",
			tagFilters:     []tuido.Tag{},
			expectedCount:  4,
			expectedItems:  []int{0, 1, 2, 3},
		},
		{
			name:           "filter by project=work",
			tagFilters:     []tuido.Tag{tuido.NewTag("project=work")},
			expectedCount:  2,
			expectedItems:  []int{0, 2},
		},
		{
			name:           "filter by project (any value)",
			tagFilters:     []tuido.Tag{tuido.NewTag("project")},
			expectedCount:  3,
			expectedItems:  []int{0, 1, 2},
		},
		{
			name:           "filter by active tag",
			tagFilters:     []tuido.Tag{tuido.NewTag("active")},
			expectedCount:  2,
			expectedItems:  []int{2, 3},
		},
		{
			name:           "filter by multiple tags (AND)",
			tagFilters:     []tuido.Tag{tuido.NewTag("project=work"), tuido.NewTag("active")},
			expectedCount:  1,
			expectedItems:  []int{2},
		},
		{
			name:           "filter with no matches",
			tagFilters:     []tuido.Tag{tuido.NewTag("nonexistent")},
			expectedCount:  0,
			expectedItems:  []int{},
		},
		// New tests for starts-with matching
		{
			name:           "partial filter by single char 'd' (matches due)",
			tagFilters:     []tuido.Tag{tuido.NewTag("d")},
			expectedCount:  2,
			expectedItems:  []int{0, 1}, // both have #due tags
		},
		{
			name:           "partial filter by 'proj' (matches project)",
			tagFilters:     []tuido.Tag{tuido.NewTag("proj")},
			expectedCount:  3,
			expectedItems:  []int{0, 1, 2}, // all have #project tags
		},
		{
			name:           "partial filter by 'act' (matches active)",
			tagFilters:     []tuido.Tag{tuido.NewTag("act")},
			expectedCount:  2,
			expectedItems:  []int{2, 3}, // both have #active tags
		},
		{
			name:           "partial filter with multiple chars 'du' (matches due)",
			tagFilters:     []tuido.Tag{tuido.NewTag("du")},
			expectedCount:  2,
			expectedItems:  []int{0, 1}, // both have #due tags
		},
		{
			name:           "partial filter combination",
			tagFilters:     []tuido.Tag{tuido.NewTag("proj"), tuido.NewTag("act")},
			expectedCount:  1,
			expectedItems:  []int{2}, // only item 2 has both project and active
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testTui.applyTagFilters(items, tt.tagFilters)
			
			if len(result) != tt.expectedCount {
				t.Errorf("applyTagFilters() returned %d items, want %d", 
					len(result), tt.expectedCount)
			}
			
			// Check that the right items were returned
			for i, expectedIdx := range tt.expectedItems {
				if i >= len(result) {
					t.Errorf("Expected item %d missing from results", expectedIdx)
					continue
				}
				expectedText := items[expectedIdx].Text()
				actualText := result[i].Text()
				if actualText != expectedText {
					t.Errorf("Result item %d: got %q, want %q", 
						i, actualText, expectedText)
				}
			}
		})
	}
}

func TestTagFilteringEdgeCases(t *testing.T) {
	// Test edge cases for tag filtering
	items := []*tuido.Item{
		func() *tuido.Item { item := tuido.New("test.txt", 1, "[ ] item with no tags"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 2, "[ ] item with empty tag #"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 3, "[ ] item with malformed tag #=value"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 4, "[ ] item with special chars #tag-name_123=value.txt"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 5, "[ ] item with multiple same tags #due=2024-01-15 #due=2024-01-16"); return &item }(),
	}

	testTui := &tui{}

	tests := []struct {
		name           string
		tagFilters     []tuido.Tag
		expectedCount  int
	}{
		{
			name:           "filter for non-existent tag returns no items",
			tagFilters:     []tuido.Tag{tuido.NewTag("nonexistent")},
			expectedCount:  0,
		},
		{
			name:           "filter for special character tag name",
			tagFilters:     []tuido.Tag{tuido.NewTag("tag-name_123")},
			expectedCount:  1,
		},
		{
			name:           "filter for special character tag with value",
			tagFilters:     []tuido.Tag{tuido.NewTag("tag-name_123=value.txt")},
			expectedCount:  1,
		},
		{
			name:           "filter handles items with no tags",
			tagFilters:     []tuido.Tag{tuido.NewTag("any")},
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testTui.applyTagFilters(items, tt.tagFilters)
			
			if len(result) != tt.expectedCount {
				t.Errorf("applyTagFilters() returned %d items, want %d", 
					len(result), tt.expectedCount)
			}
		})
	}
}

func TestApplyFuzzySearch(t *testing.T) {
	// Create test items for fuzzy search
	items := []*tuido.Item{
		func() *tuido.Item { item := tuido.New("test.txt", 1, "[ ] fix critical bug in authentication"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 2, "[ ] write documentation for API"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 3, "[ ] review code changes"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 4, "[ ] update user interface"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 5, "[ ] fix minor bug in UI"); return &item }(),
	}

	testTui := &tui{}

	tests := []struct {
		name           string
		terms          []string
		expectedCount  int
		expectedItems  []int // indices of items that should match
	}{
		{
			name:           "single fuzzy term",
			terms:          []string{"bug"},
			expectedCount:  2,
			expectedItems:  []int{0, 4}, // "critical bug" and "bug in UI"
		},
		{
			name:           "fuzzy matching with typo",
			terms:          []string{"docmentation"}, // missing 'u'
			expectedCount:  1,
			expectedItems:  []int{1}, // should match "documentation"
		},
		{
			name:           "multiple terms (AND logic)",
			terms:          []string{"fix", "bug"},
			expectedCount:  2,
			expectedItems:  []int{0, 4}, // both have "fix" and "bug"
		},
		{
			name:           "no terms returns all",
			terms:          []string{},
			expectedCount:  5,
			expectedItems:  []int{0, 1, 2, 3, 4},
		},
		{
			name:           "no matches",
			terms:          []string{"nonexistent"},
			expectedCount:  0,
			expectedItems:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testTui.applyFuzzySearch(items, tt.terms)
			
			if len(result) != tt.expectedCount {
				t.Errorf("applyFuzzySearch() returned %d items, want %d", 
					len(result), tt.expectedCount)
			}
			
			// Check that we got the expected items (order may vary due to fuzzy ranking)
			if len(tt.expectedItems) > 0 {
				expectedTexts := make(map[string]bool)
				for _, idx := range tt.expectedItems {
					expectedTexts[items[idx].Text()] = true
				}
				
				for _, item := range result {
					if !expectedTexts[item.Text()] {
						t.Errorf("Unexpected item in results: %q", item.Text())
					}
				}
			}
		})
	}
}

func TestMultiModalFiltering(t *testing.T) {
	// Create test items with both tags and text for comprehensive testing
	items := []*tuido.Item{
		func() *tuido.Item { item := tuido.New("test.txt", 1, "[ ] fix authentication bug #due=2024-01-15 #project=security"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 2, "[ ] write API documentation #due=2024-01-16 #project=docs"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 3, "[ ] fix UI bug in login page #project=security #active"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 4, "[ ] review security code changes #project=security"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 5, "[ ] update user interface design #project=design #active"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 6, "[ ] write unit tests #project=testing"); return &item }(),
	}

	filter := textinput.New()
	testTui := &tui{
		renderSelection: items,
		filter:         filter,
	}

	tests := []struct {
		name           string
		query          string
		expectedCount  int
		description    string
	}{
		{
			name:           "tag filter only",
			query:          "#project=security",
			expectedCount:  3,
			description:    "Items with project=security tag",
		},
		{
			name:           "tag filter with any value",
			query:          "#project",
			expectedCount:  6,
			description:    "All items have project tags",
		},
		{
			name:           "fuzzy search only",
			query:          "bug",
			expectedCount:  2,
			description:    "Items mentioning bugs",
		},
		{
			name:           "tag + fuzzy combination",
			query:          "#project=security fix",
			expectedCount:  2,
			description:    "Security project items mentioning 'fix'",
		},
		{
			name:           "multiple tags + fuzzy",
			query:          "#project=security #active interface",
			expectedCount:  0,
			description:    "Security + active items mentioning interface (none match)",
		},
		{
			name:           "multiple tags + fuzzy match",
			query:          "#project=security bug",
			expectedCount:  2,
			description:    "Security items mentioning bugs",
		},
		{
			name:           "fuzzy with typo + tag",
			query:          "#active interfce", // missing 'a' in interface
			expectedCount:  1,
			description:    "Active items fuzzy matching 'interface'",
		},
		{
			name:           "no filters",
			query:          "",
			expectedCount:  6,
			description:    "Empty query returns all items",
		},
		{
			name:           "tag with no matches",
			query:          "#nonexistent",
			expectedCount:  0,
			description:    "Non-existent tag returns no items",
		},
		{
			name:           "multiple fuzzy terms",
			query:          "write documentation",
			expectedCount:  1,
			description:    "Items matching both 'write' and 'documentation'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset renderSelection for each test
			testTui.renderSelection = items
			testTui.filter.SetValue(tt.query)
			
			// Apply the filter
			testTui.applyFilter()
			
			result := testTui.renderSelection
			if len(result) != tt.expectedCount {
				t.Errorf("Multi-modal filter with query %q returned %d items, want %d\nDescription: %s", 
					tt.query, len(result), tt.expectedCount, tt.description)
				
				// Debug output
				t.Logf("Items found:")
				for i, item := range result {
					t.Logf("  %d: %s", i, item.Text())
				}
			}
		})
	}
}

func TestEndToEndMultiModalFiltering(t *testing.T) {
	// Comprehensive end-to-end test for multi-modal filtering
	items := []*tuido.Item{
		func() *tuido.Item { item := tuido.New("test.txt", 1, "[ ] fix critical authentication bug #due=2024-01-15 #project=security #urgent"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 2, "[ ] write comprehensive API documentation #due=2024-01-16 #project=docs"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 3, "[ ] fix minor UI bug in login page #project=security #active"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 4, "[ ] review security code changes #project=security #urgent"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 5, "[ ] update user interface design #project=design #active"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 6, "[ ] write unit tests for authentication #project=testing #urgent"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 7, "[ ] fix database connection issue #project=backend"); return &item }(),
		func() *tuido.Item { item := tuido.New("test.txt", 8, "[ ] implement new authentication flow #project=security #due=2024-01-20"); return &item }(),
	}

	filter := textinput.New()
	testTui := &tui{
		renderSelection: items,
		filter:         filter,
	}

	tests := []struct {
		name           string
		query          string
		expectedItems  []string // exact text matches expected
	}{
		{
			name:  "complex tag + fuzzy query",
			query: "#project=security #urgent authentication",
			expectedItems: []string{
				"fix critical authentication bug #due=2024-01-15 #project=security #urgent",
			},
		},
		{
			name:  "tag with any value + fuzzy",
			query: "#project bug",
			expectedItems: []string{
				"fix critical authentication bug #due=2024-01-15 #project=security #urgent",
				"fix minor UI bug in login page #project=security #active",
			},
		},
		{
			name:  "multiple tags + multiple fuzzy terms",
			query: "#project=security fix",
			expectedItems: []string{
				"fix critical authentication bug #due=2024-01-15 #project=security #urgent",
				"fix minor UI bug in login page #project=security #active",
			},
		},
		{
			name:  "fuzzy with typo + tag filter",
			query: "#urgent athentication", // missing 'u' in authentication
			expectedItems: []string{
				"fix critical authentication bug #due=2024-01-15 #project=security #urgent",
				"write unit tests for authentication #project=testing #urgent",
			},
		},
		{
			name:  "tag only filter",
			query: "#active",
			expectedItems: []string{
				"fix minor UI bug in login page #project=security #active",
				"update user interface design #project=design #active",
			},
		},
		{
			name:  "fuzzy only search",
			query: "documentation",
			expectedItems: []string{
				"write comprehensive API documentation #due=2024-01-16 #project=docs",
			},
		},
		{
			name:           "no matches - strict filter",
			query:          "#nonexistent exact",
			expectedItems:  []string{},
		},
		{
			name:  "empty query returns all",
			query: "",
			expectedItems: []string{
				"fix critical authentication bug #due=2024-01-15 #project=security #urgent",
				"write comprehensive API documentation #due=2024-01-16 #project=docs",
				"fix minor UI bug in login page #project=security #active",
				"review security code changes #project=security #urgent",
				"update user interface design #project=design #active",
				"write unit tests for authentication #project=testing #urgent",
				"fix database connection issue #project=backend",
				"implement new authentication flow #project=security #due=2024-01-20",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state for each test
			testTui.renderSelection = items
			testTui.filter.SetValue(tt.query)
			
			// Apply the multi-modal filter
			testTui.applyFilter()
			
			result := testTui.renderSelection
			
			// Check count
			if len(result) != len(tt.expectedItems) {
				t.Errorf("Query %q returned %d items, want %d", 
					tt.query, len(result), len(tt.expectedItems))
			}
			
			// Check exact matches (convert to map for easier comparison)
			expectedMap := make(map[string]bool)
			for _, expected := range tt.expectedItems {
				expectedMap[expected] = true
			}
			
			actualMap := make(map[string]bool)
			for _, item := range result {
				actualMap[item.Text()] = true
			}
			
			// Verify all expected items are present
			for expected := range expectedMap {
				if !actualMap[expected] {
					t.Errorf("Expected item missing: %q", expected)
				}
			}
			
			// Verify no unexpected items
			for actual := range actualMap {
				if !expectedMap[actual] {
					t.Errorf("Unexpected item found: %q", actual)
				}
			}
		})
	}
}

func TestParseFilterQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected filterQuery
	}{
		{
			name:  "empty input",
			input: "",
			expected: filterQuery{
				tagFilters: nil,
				fuzzyTerms: nil,
			},
		},
		{
			name:  "only fuzzy terms",
			input: "fix bug meeting",
			expected: filterQuery{
				tagFilters: nil,
				fuzzyTerms: []string{"fix", "bug", "meeting"},
			},
		},
		{
			name:  "only tag filters",
			input: "#due #active",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("due"),
					tuido.NewTag("active"),
				},
				fuzzyTerms: nil,
			},
		},
		{
			name:  "only tag filters with values",
			input: "#due=2024-01-15 #project=work",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("due=2024-01-15"),
					tuido.NewTag("project=work"),
				},
				fuzzyTerms: nil,
			},
		},
		{
			name:  "mixed tags and fuzzy",
			input: "#due bug fix",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("due"),
				},
				fuzzyTerms: []string{"bug", "fix"},
			},
		},
		{
			name:  "complex mixed query",
			input: "#project=work #active meeting important",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("project=work"),
					tuido.NewTag("active"),
				},
				fuzzyTerms: []string{"meeting", "important"},
			},
		},
		{
			name:  "ignore standalone hash",
			input: "# fix bug",
			expected: filterQuery{
				tagFilters: nil,
				fuzzyTerms: []string{"#", "fix", "bug"},
			},
		},
		{
			name:  "hash in middle of word treated as fuzzy",
			input: "bug#fix meeting",
			expected: filterQuery{
				tagFilters: nil,
				fuzzyTerms: []string{"bug#fix", "meeting"},
			},
		},
		{
			name:  "multiple spaces and tabs",
			input: "  #due   bug    #active  ",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("due"),
					tuido.NewTag("active"),
				},
				fuzzyTerms: []string{"bug"},
			},
		},
		{
			name:  "tag with empty value",
			input: "#project= meeting",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("project="),
				},
				fuzzyTerms: []string{"meeting"},
			},
		},
		{
			name:  "special characters in tag value",
			input: "#note=check-this_out.md #active",
			expected: filterQuery{
				tagFilters: []tuido.Tag{
					tuido.NewTag("note=check-this_out.md"),
					tuido.NewTag("active"),
				},
				fuzzyTerms: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFilterQuery(tt.input)
			
			// Compare tag filters
			if len(result.tagFilters) != len(tt.expected.tagFilters) {
				t.Errorf("parseFilterQuery(%s) tagFilters length = %d, want %d", 
					tt.input, len(result.tagFilters), len(tt.expected.tagFilters))
			} else {
				for i, filter := range result.tagFilters {
					if filter.Name() != tt.expected.tagFilters[i].Name() || filter.String() != tt.expected.tagFilters[i].String() {
						t.Errorf("parseFilterQuery(%s) tagFilters[%d] = %+v, want %+v", 
							tt.input, i, filter, tt.expected.tagFilters[i])
					}
				}
			}
			
			// Compare fuzzy terms
			if !reflect.DeepEqual(result.fuzzyTerms, tt.expected.fuzzyTerms) {
				t.Errorf("parseFilterQuery(%s) fuzzyTerms = %+v, want %+v", 
					tt.input, result.fuzzyTerms, tt.expected.fuzzyTerms)
			}
		})
	}
}