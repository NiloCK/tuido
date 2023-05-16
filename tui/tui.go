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
)

func Run() {
	wdStr, err := os.Getwd() // [ ] only from cli flag? YES! or... follow .gitignore

	if err != nil {
		panic(err)
	}

	adoptConfigSettings(filepath.Join(wdStr, ".tuido"))
	// [ ] read cli flags for added extensions / extension specificity

	files := []string{}

	wtStat, err := os.Stat(runConfig.writeto)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if wtStat.IsDir() {
		files = append(files, getFiles(runConfig.writeto, runConfig.extensions)...)
	}

	// [ ] replace with subdir check #active=2022-05-26 #zzz=2
	if wdStr != runConfig.writeto {
		wdFiles := getFiles(wdStr, runConfig.extensions)
		files = append(files, wdFiles...)
	}

	items := []*tuido.Item{}
	for _, f := range files {
		items = append(items, getItems(f)...)
	}

	sortItems(items)

	prog := tea.NewProgram(newTUI(items, runConfig), tea.WithAltScreen())

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
	filter.Placeholder = "filter by #tag. press /"

	itemEditor := textinput.New()
	itemEditor.Prompt = ">>>"

	return tui{
		config:          cfg,
		err:             nil,
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
)

type tui struct {
	config config
	err    error

	items       []*tuido.Item
	itemsFilter itemType

	renderSelection []*tuido.Item
	selection       int
	pages           int
	currentPage     int

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

	nag  nagScreen
	peek peekScreen

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
	t.pomoTimeRemaining = int(setTime * 60)
	t.pomoTimeSet = t.pomoTimeRemaining

	if err != nil {
		t.err = err
		fmt.Println(err)
	}
}

func (t *tui) setPeekMode() tea.Cmd {
	t.mode = peek
	t.peek = peekScreen{*t.currentSelection()}
	return nil
}

func (t *tui) setEditMode() tea.Cmd {
	t.mode = edit
	t.itemEditor.SetValue(t.currentSelection().Text())
	t.itemEditor.CursorEnd()
	t.itemEditor.Focus()
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

	t.applyTagFilters()
	sortItems(t.renderSelection)
	// ensure the previous selection value is still in range
	t.setSelection(t.selection)
}

func (t *tui) applyTagFilters() {
	filterTags := tuido.Tags(t.filter.Value())
	if len(filterTags) != 0 {

		filtered := []*tuido.Item{}

		for _, item := range t.renderSelection {
			itemTags := item.Tags()

			for _, iTag := range itemTags {
				for _, fTag := range filterTags {
					// [ ] should not use the prefix when a tag is "complete" (followed by a space) in the prompt
					if strings.HasPrefix(iTag.Name(), fTag.Name()) {
						filtered = append(filtered, item)
						continue
					}
				}
			}
		}

		t.renderSelection = filtered
	}
}

func (t tui) Init() tea.Cmd { return tick() }

func getItems(file string) []*tuido.Item {
	items := []*tuido.Item{}

	f, err := os.Open(file)
	defer f.Close()

	if err != nil {
		panic(err)
	}

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

func getFiles(wd string, extensions []string) []string {

	files := []string{}
	filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
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

func sortItems(items []*tuido.Item) {
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
			return true
		} else if x == nil && y != nil {
			return false
		} else if x != nil && y == nil {
			return true
		} else {
			return x.Before(*y)
		}
	})
}
