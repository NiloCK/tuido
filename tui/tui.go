package tui

import (
	"bufio"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nilock/tuido/tuido"
)

func Run() {
	wdStr, err := os.Getwd()

	extensions := []string{
		"md",
		"txt",
		"xit",
	}
	// todo: flag for added extensions / extension specificity

	if err != nil {
		panic(err)
	}

	files := getFiles(wdStr, extensions)

	items := []*tuido.Item{}
	for _, f := range files {
		items = append(items, getItems(f)...)
	}

	prog := tea.NewProgram(newTUI(items), tea.WithAltScreen())

	if err := prog.Start(); err != nil {
		panic(err)
	}
}

type itemType string

const (
	todo itemType = "todo"
	done itemType = "done"
)

func init() {
	// home := os.Getenv("HOME")
	// tdpath := path.Join(home, ".tuido")

	rand.Seed(time.Now().Unix()) // a fresh set of tag colors on each run. Spice of life.
}

func newTUI(items []*tuido.Item) tui {
	// the search bar:
	filter := textinput.New()
	filter.Placeholder = "filter by #tag. press /"

	return tui{
		items:           items,
		renderSelection: nil,
		itemsFilter:     todo,
		mode:            navigation,
		selection:       0,
		filter:          filter,
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
	help
)

type tui struct {
	items       []*tuido.Item
	itemsFilter itemType

	renderSelection []*tuido.Item
	selection       int
	pages           int
	currentPage     int

	mode mode

	filter textinput.Model

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

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if t.mode == help {
		if _, ok := msg.(tea.KeyMsg); ok {
			t.mode = navigation
			return t, nil
		}
	}

	t.populateRenderSelection()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if t.filter.Focused() {
			k := msg.String()
			if k == "esc" ||
				k == "tab" ||
				k == "down" {
				t.filter.Blur()
			} else {
				var cmd tea.Cmd
				t.filter, cmd = t.filter.Update(msg)

				return t, cmd
			}
		}

		switch msg.String() {
		case "up":
			t.setSelection(t.selection - 1)
		case "down":
			t.setSelection(t.selection + 1)
		case "pgdown": // [ ] these paging functions are not "accurate" #ui #polish
			t.setSelection(t.selection + (len(t.renderSelection) / (t.h - 6)))
		case "pgup":
			t.setSelection(t.selection - (len(t.renderSelection) / (t.h - 6)))
		case "tab":
			t.tab()
		case "x":
			t.currentSelection().SetStatus(tuido.Checked)
		case "-":
			t.currentSelection().SetStatus(tuido.Obsolete)
		case "~":
			t.currentSelection().SetStatus(tuido.Obsolete)
		case "s":
			t.currentSelection().SetStatus(tuido.Obsolete)
		case "@":
			t.currentSelection().SetStatus(tuido.Ongoing)
		case "a":
			t.currentSelection().SetStatus(tuido.Ongoing)
		case " ":
			t.currentSelection().SetStatus(tuido.Open)
		case "/":
			t.filter.Focus()
		case "?":
			t.mode = help
		case "q":
			return t, tea.Quit
		}

	case tea.WindowSizeMsg:
		t.h = msg.Height
		t.w = msg.Width
	}
	return t, nil
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

		if d.IsDir() { // apply .tuido config if it exists

			configPath := filepath.Join(path, ".tuido")

			if config, err := os.Open(configPath); err == nil {
				extensions = append(extensions, parseConfig(config)...)
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

func parseConfig(cfg *os.File) (extensions []string) {
	buf := make([]byte, 1024)

	if n, err := cfg.Read(buf); err == nil {
		cStr := string(buf[:n])
		fmt.Println(buf)
		split := strings.Split(cStr, "=")

		if split[0] == "extensions" {
			return strings.Split(split[1], ",")
		}
	}

	return nil
}
