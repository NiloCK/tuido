package tui

import (
	"bufio"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nilock/tuido/tuido"
	"github.com/nilock/tuido/utils"
	walkrepo "github.com/nilock/walk-repo"
	"github.com/sahilm/fuzzy"
)

// Filter query structures for multi-modal tag + fuzzy search
type filterQuery struct {
	tagFilters []tuido.Tag
	fuzzyTerms []string
}

func Run() { run("") }

// RunFocused launches the TUI scoped to a single file (filezoom). Every item
// in the file is shown in file order; group collapse is disabled.
func RunFocused(file string) { run(file) }

func run(focus string) {
	// In focus mode, load only the focused file's items - this guarantees the
	// focused path matches the items' File() exactly, sidestepping any
	// relative/absolute path mismatch.
	if focus != "" {
		items := GetItems(focus)
		tui := newTUI(items, runConfig)
		tui.focused = focus

		prog := tea.NewProgram(tui, tea.WithAltScreen())
		if err := prog.Start(); err != nil {
			panic(err)
		}
		return
	}

	wrkdirStr, err := os.Getwd() // [ ] only from cli flag? YES! or... follow .gitignore

	if err != nil {
		panic(err)
	}

	// [ ] read cli flags for added extensions / extension specificity

	files := make(map[string]struct{})

	wtStat, err := os.Stat(runConfig.writeto)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if wtStat.IsDir() {
		writeDirFiles := GetFiles(runConfig.writeto, runConfig.extensions)
		for _, f := range writeDirFiles {
			files[f] = struct{}{}
		}
	}

	// [ ] replace with subdir check #active=2022-05-26 #zzz=2
	if wrkdirStr != runConfig.writeto {
		wdFiles := GetFiles(wrkdirStr, runConfig.extensions)
		for _, f := range wdFiles {
			files[f] = struct{}{}
		}
	}

	items := []*tuido.Item{}
	for f := range files {
		items = append(items, GetItems(f)...)
	}

	SortItems(items)

	tui := newTUI(items, runConfig)
	tui.houseKeeping()

	prog := tea.NewProgram(tui, tea.WithAltScreen())

	if err := prog.Start(); err != nil {
		panic(err)
	}
}

type itemType string

const (
	todo itemType = "todo"
	done itemType = "done"
)

func newTUI(items []*tuido.Item, cfg config) tui {
	// the search bar:
	filter := textinput.New()
	filter.Placeholder = "filter (press /)"

	itemEditor := textinput.New()
	itemEditor.Prompt = ">>>"

	return tui{
		config:          cfg,
		err:             nil,
		notifs:          []string{},
		items:           items,
		renderSelection: nil,
		itemsFilter:     todo,
		mode:            navigation,
		selection:       0,
		pomoEditor:      textinput.New(),
		filter:          filter,
		itemEditor:      itemEditor,
		tagColors:       populateTagColorStyles(items),
		h:               0,
		w:               0,
	}
}

func (t *tui) houseKeeping() {
	local := utils.Version()
	if local == "dev" {
		// Unversioned local build; no meaningful release to compare against.
		return
	}
	curent := utils.LatestVersion()

	if local != curent {
		t.notifs = append(t.notifs,
			fmt.Sprintf("New version available: %s. Currently running %s. \nPress 'u' to upgrade, or visit %s for more info",
				curent, local, utils.ReleaseURL))
	}

}

// populateTagColorStyles returns a coloring style for
// each #tag that exists in the list of items.
func populateTagColorStyles(items []*tuido.Item) map[string]lg.Style {
	// [ ] this should be recalculated / shifted when new tags are added
	// [ ] audit: results in UI suggest a bug. Colors seem clustered. ##active=2022-05-26 ##zzz=2 #active=2022-05-25 #zzz=1
	var tags []tuido.Tag
	for _, item := range items {
		tags = append(tags, item.Tags()...)
	}

	tagColors := map[string]lg.Style{}
	interval := 360.0 / float64(len(tags))
	offset := rand.Float64() * 360

	for i, tag := range tags {
		hue := int(offset+float64(i)*interval) % 360
		tagColors[tag.Name()] = lg.NewStyle().
			Foreground(
				lg.Color(
					colorful.Hcl(float64(hue), .9, 0.85).Clamped().Hex(),
				),
			)
	}
	return tagColors
}

type mode int

const (
	navigation mode = iota
	filter
	edit
	help
	pomo
	nag
	peek
	configViewer
	upgrade
)

type tui struct {
	config config
	err    error

	notifs []string

	items       []*tuido.Item
	itemsFilter itemType

	renderSelection []*tuido.Item
	selection       int
	pages           int
	currentPage     int

	// focused, when non-empty, scopes the list to a single file's items
	// (file-scoped "focus" / filezoom). In focus mode, that file's control
	// item is not collapsed - all of its items are shown in file order.
	focused string

	mode mode

	filter     textinput.Model
	itemEditor textinput.Model

	// pomoEditor is the textinput.Model for the pomo clock
	pomoEditor textinput.Model
	// pomoTimer is the ticker that decrements the pomo clock
	pomoTimer time.Ticker
	// pomoTimeRemaining is the time remaining in seconds
	pomoTimeRemaining int
	// pomoTimeSet is the original time set by the user
	pomoTimeSet int

	nag     nagScreen
	peek    peekScreen
	upgrade upgradeModel

	tagColors map[string]lg.Style

	// height of the window
	h int
	// width of the window
	w int
}

func (t *tui) setSelection(s int) {
	s = min(s, len(t.renderSelection)-1)
	s = max(s, 0)

	t.selection = s
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (t *tui) setPomoMode() tea.Cmd {
	t.mode = pomo

	t.pomoEditor.Focus()
	t.pomoEditor.SetValue("")
	t.pomoTimer.Stop()

	return nil
}

func (t *tui) startPomo() {
	if t.pomoEditor.Value() == "" {
		return
	}

	var err error
	setTime, err := strconv.ParseFloat(t.pomoEditor.Value(), 64)
	t.pomoTimeSet = int(setTime * 60)
	t.pomoTimeRemaining = t.pomoTimeSet

	if err != nil {
		t.err = err
		fmt.Println(err)
	}
}

func (t *tui) setPeekMode() tea.Cmd {
	t.mode = peek

	if t.currentSelection() != nil {
		t.peek = peekScreen{*t.currentSelection()}
	} else if len(t.items) != 0 {
		t.peek = peekScreen{*t.items[0]}
	}

	return nil
}

func (t *tui) setEditMode() tea.Cmd {
	if t.currentSelection() != nil {
		t.mode = edit
		t.itemEditor.SetValue(t.currentSelection().Text())
		t.itemEditor.CursorEnd()
		t.itemEditor.Focus()
	}
	return nil
}

// tab cycles the view between todos and dones.
func (t *tui) tab() {

	if t.itemsFilter == todo {
		t.itemsFilter = done
	} else if t.itemsFilter == done {
		t.itemsFilter = todo
	}

	t.populateRenderSelection()
}

func (t *tui) currentSelection() *tuido.Item {
	if len(t.renderSelection) == 0 {
		t.populateRenderSelection()
		return nil
	}
	t.setSelection(t.selection)
	return t.renderSelection[t.selection]
}

// populateRenderSelection pulls appropriate items from
// the global items slice into the renderSelection slice
// based on their status and the current selected view.
func (t *tui) populateRenderSelection() {
	t.renderSelection = []*tuido.Item{}

	// In focus mode, scope to the focused file and show every item in it
	// (in file order), bypassing status filtering and group collapse.
	if t.focused != "" {
		for _, i := range t.items {
			if i.File() == t.focused {
				t.renderSelection = append(t.renderSelection, i)
			}
		}
		t.applyFilter()
		if len(t.filter.Value()) == 0 {
			sort.SliceStable(t.renderSelection, func(a, b int) bool {
				return t.renderSelection[a].Line() < t.renderSelection[b].Line()
			})
		}
		t.setSelection(t.selection)
		return
	}

	if t.itemsFilter == todo {
		for _, i := range t.items {
			if (i.Satus() == tuido.Ongoing || i.Satus() == tuido.Open) &&
				i.Active() {
				t.renderSelection = append(t.renderSelection, i)
			}
		}
	}

	if t.itemsFilter == done {
		for _, i := range t.items {
			if i.Satus() == tuido.Checked || i.Satus() == tuido.Obsolete {
				t.renderSelection = append(t.renderSelection, i)
			}
		}
	}

	// collapse file-scoped groups (##file) to their control item
	t.renderSelection = tuido.CollapseFileScoped(t.renderSelection)

	t.applyFilter()

	// Only sort if no filter is active - preserve fuzzy search ranking
	if len(t.filter.Value()) == 0 {
		SortItems(t.renderSelection)
	}

	// ensure the previous selection value is still in range
	t.setSelection(t.selection)
}

// enterFocus scopes the view to a single file (filezoom).
func (t *tui) enterFocus(file string) {
	t.focused = file
	t.selection = 0
	t.populateRenderSelection()
}

// exitFocus returns from filezoom to the aggregate view.
func (t *tui) exitFocus() {
	t.focused = ""
	t.selection = 0
	t.populateRenderSelection()
}

func (t *tui) applyFilter() {
	query := t.filter.Value()
	if len(query) == 0 {
		return
	}

	// Parse the query into tag filters and fuzzy terms
	filterQuery := parseFilterQuery(query)

	// Start with current selection
	currentSelection := t.renderSelection

	// Apply hard tag filtering first
	if len(filterQuery.tagFilters) > 0 {
		currentSelection = t.applyTagFilters(currentSelection, filterQuery.tagFilters)
	}

	// Apply fuzzy search on remaining terms
	if len(filterQuery.fuzzyTerms) > 0 {
		currentSelection = t.applyFuzzySearch(currentSelection, filterQuery.fuzzyTerms)
	}

	t.renderSelection = currentSelection
}

// parseFilterQuery parses filter input into tag filters and fuzzy search terms
func parseFilterQuery(input string) filterQuery {
	tokens := strings.Fields(input)
	var tagFilters []tuido.Tag
	var fuzzyTerms []string

	for _, token := range tokens {
		if strings.HasPrefix(token, "#") && len(token) > 1 {
			tagFilters = append(tagFilters, parseTagFilter(token))
		} else {
			fuzzyTerms = append(fuzzyTerms, token)
		}
	}

	return filterQuery{tagFilters, fuzzyTerms}
}

// parseTagFilter parses a tag filter token like "#tag" or "#tag=value"
func parseTagFilter(token string) tuido.Tag {
	// Remove leading # and use existing NewTag function from tuido package
	return tuido.NewTag(token[1:])
}

// applyFuzzySearch applies fuzzy search to items using the provided terms
func (t *tui) applyFuzzySearch(items []*tuido.Item, terms []string) []*tuido.Item {
	currentSelection := items

	// For each fuzzy term, filter the current selection
	for _, term := range terms {
		if len(currentSelection) == 0 {
			break
		}

		// Prepare search targets for fuzzy matching
		var searchTargets []string
		for _, item := range currentSelection {
			searchTargets = append(searchTargets, item.Text())
		}

		// Perform fuzzy search on current term
		matches := fuzzy.Find(term, searchTargets)

		// Update selection to only include fuzzy matches
		newSelection := make([]*tuido.Item, len(matches))
		for i, match := range matches {
			newSelection[i] = currentSelection[match.Index]
		}
		currentSelection = newSelection
	}

	return currentSelection
}

// applyTagFilters filters items based on tag criteria (hard filtering)
func (t *tui) applyTagFilters(items []*tuido.Item, tagFilters []tuido.Tag) []*tuido.Item {
	if len(tagFilters) == 0 {
		return items
	}

	var filtered []*tuido.Item

	for _, item := range items {
		if itemMatchesAllTags(item, tagFilters) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// itemMatchesAllTags checks if an item has all required tags (AND logic)
func itemMatchesAllTags(item *tuido.Item, tagFilters []tuido.Tag) bool {
	itemTags := item.Tags()

	for _, filter := range tagFilters {
		if !itemHasTag(itemTags, filter) {
			return false // AND logic - all tags must match
		}
	}

	return true
}

// itemHasTag checks if an item has a specific tag (with optional value matching)
func itemHasTag(itemTags []tuido.Tag, filter tuido.Tag) bool {
	for _, tag := range itemTags {
		// Use starts-with matching for tag names to enable partial matching
		if strings.HasPrefix(tag.Name(), filter.Name()) {
			// If filter has no value (just tag name), any value matches
			if filter.String() == filter.Name() {
				return true
			}
			// If filter has value, tag name must start with filter name AND value must match exactly
			if tag.String() == filter.String() {
				return true
			}
		}
	}
	return false
}

func (t tui) Init() tea.Cmd { return tick() }

func GetItems(file string) []*tuido.Item {
	items := []*tuido.Item{}

	if f, err := os.Open(file); err != nil {
		f.Close()
		return items
	} else {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		line := 1
		for scanner.Scan() {
			if tuido.IsTuido(scanner.Text()) {
				item := tuido.New(file, line, scanner.Text())
				items = append(items, &item)
			}
			line++
		}

		return items
	}
}

func GetFiles(wd string, extensions []string) []string {

	files := []string{}

	walkrepo.WalkRepo(wd, func(path string, d fs.FileInfo, err error) error {
		// apply .tuido configured extensions if they exist, but do not
		// read a configured writeto. writeto is decided by the root
		// working directory or user config
		if d.IsDir() {
			cfg := parseConfigIfExists(filepath.Join(path, ".tuido"))
			if cfg != nil {
				extensions = cfg.extensions
			}
		}

		for _, suffix := range extensions {

			if strings.HasSuffix(
				strings.ToLower(path),
				suffix,
			) {
				files = append(files, path)
			}

		}
		return nil
	})
	return files
}

func SortItems(items []*tuido.Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Importance() > items[j].Importance() {
			return true
		}
		if items[i].Importance() < items[j].Importance() {
			return false
		}

		x := items[i].Due()
		y := items[j].Due()

		if x == nil && y == nil {
			return true // no swap, leave original order (grouped by file, order of appearance)
		} else if x == nil && y != nil {
			return false
		} else if x != nil && y == nil {
			return true
		} else {
			return x.Before(*y)
		}
	})
}
