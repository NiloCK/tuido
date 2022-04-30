package tui

import (
	"bufio"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type view string

const (
	todo view = "todo"
	done view = "done"
)

func init() {
	rand.Seed(time.Now().Unix()) // a fresh set of tag colors on each run. Spice of life.
}

func newTUI(items []*tuido.Item) tui {
	// the search bar:
	filter := textinput.New()
	filter.Placeholder = "filter by #tag. press /"

	return tui{
		items:           items,
		renderSelection: nil,
		view:            todo,
		selection:       0,
		filter:          filter,
		tagColors:       populateTagColorStyles(items),
		h:               0,
		w:               0,
	}
}

// populateTagColorStyles returns a coloring style for
// each #tag that exists in the list of items.
func populateTagColorStyles(items []*tuido.Item) map[string]lipgloss.Style {
	var tags []string
	for _, item := range items {
		tags = append(tags, item.Tags()...)
	}

	tagColors := map[string]lipgloss.Style{}
	interval := 360.0 / float64(len(tags))
	offset := rand.Float64() * 360

	for i, tag := range tags {
		hue := int(offset+float64(i)*interval) % 360
		tagColors[tag] = lipgloss.NewStyle().
			Foreground(
				lipgloss.Color(
					colorful.Hcl(float64(hue), .9, 0.85).Clamped().Hex(),
				),
			)
	}
	return tagColors
}

type tui struct {
	items           []*tuido.Item
	renderSelection []*tuido.Item
	view            view
	selection       int

	filter textinput.Model

	tagColors map[string]lipgloss.Style

	// height of the window
	h int
	// width of the window
	w int
}

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if t.selection > 0 {
				t.selection--
			}
		case "down":
			if t.selection+1 < len(t.renderSelection) {
				t.selection++
			}
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

	if t.view == todo {
		t.view = done
	} else if t.view == done {
		t.view = todo
	}

	t.populateRenderSelection()
}

func (t *tui) currentSelection() *tuido.Item {
	if len(t.renderSelection) == 0 {
		t.populateRenderSelection()
	}
	return t.renderSelection[t.selection]
}

// populateRenderSelection pulls appropriate items from
// the global items slice into the renderSelection slice
// based on their status and the current selected view.
func (t *tui) populateRenderSelection() {
	t.renderSelection = []*tuido.Item{}

	if t.view == todo {
		for _, i := range t.items {
			if i.Satus() == tuido.Ongoing || i.Satus() == tuido.Open {
				t.renderSelection = append(t.renderSelection, i)
			}
		}
	}

	if t.view == done {
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

	if t.selection+1 >= len(t.renderSelection) {
		t.selection = len(t.renderSelection) - 1
	}
}

func (t tui) Init() tea.Cmd { return textinput.Blink }

func getItems(file string) []*tuido.Item {
	prefixes := []string{"[ ]", "[@]", "[x]", "[~]", "[?]"}
	items := []*tuido.Item{}

	f, err := os.Open(file)
	defer f.Close()

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	line := 1
	for scanner.Scan() {
		for _, prefix := range prefixes {
			if strings.HasPrefix(scanner.Text(), prefix) {
				item := tuido.New(file, line, scanner.Text())
				items = append(items, &item)
			}
		}
		line++
	}

	return items
}

func getFiles(wd string, extensions []string) []string {
	files := []string{}
	filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
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
