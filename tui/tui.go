package tui

import (
	"bufio"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

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

	adoptConfigSettings("~/.tuido/.config")
	adoptConfigSettings(filepath.Join(wdStr, ".tuido"))
	// [ ] read cli flags for added extensions / extension specificity

	wdFiles := getFiles(wdStr, runConfig.extensions)
	cfgFiles := getFiles(runConfig.writeto, runConfig.extensions)

	files := append(wdFiles, cfgFiles...)

	items := []*tuido.Item{}
	for _, f := range files {
		items = append(items, getItems(f)...)
	}

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
		items:           items,
		renderSelection: nil,
		itemsFilter:     todo,
		mode:            navigation,
		selection:       0,
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
	var tags []string
	for _, item := range items {
		tags = append(tags, item.Tags()...)
	}

	tagColors := map[string]lg.Style{}
	interval := 360.0 / float64(len(tags))
	offset := rand.Float64() * 360

	for i, tag := range tags {
		hue := int(offset+float64(i)*interval) % 360
		tagColors[tag] = lg.NewStyle().
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
)

type tui struct {
	config config

	items       []*tuido.Item
	itemsFilter itemType

	renderSelection []*tuido.Item
	selection       int
	pages           int
	currentPage     int

	mode mode

	filter     textinput.Model
	itemEditor textinput.Model

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
			if i.Satus() == tuido.Ongoing || i.Satus() == tuido.Open {
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

	filterTags := tuido.Tags(t.filter.Value())
	if len(filterTags) != 0 {

		filtered := []*tuido.Item{}

		for _, item := range t.renderSelection {
			itemTags := item.Tags()

			for _, iTag := range itemTags {
				for _, fTag := range filterTags {
					// [ ] should not use the prefix when a tag is "complete" (followed by a space) in the prompt
					if strings.HasPrefix(iTag, fTag) {
						filtered = append(filtered, item)
						continue
					}
				}
			}
		}

		t.renderSelection = filtered
	}

	// ensure the previous selection value is still in range
	t.setSelection(t.selection)
}

func (t tui) Init() tea.Cmd { return textinput.Blink }

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
