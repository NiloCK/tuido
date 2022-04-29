package tui

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	prog := tea.NewProgram(tuido, tea.WithAltScreen())

	if err := prog.Start(); err != nil {
		panic(err)
	}
}

type view string

const (
	todo view = "todo"
	done view = "done"
)

type tui struct {
	items           []*tuido.Item
	renderSelection []*tuido.Item
	view            view
	selection       int
	// height of the window
	h int
	// width of the window
	w int
}

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	t.populateRenderSelection()
	switch msg := msg.(type) {
	case tea.KeyMsg:
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

	if t.selection+1 >= len(t.renderSelection) {
		t.selection = len(t.renderSelection) - 1
	}
}

func (t tui) Init() tea.Cmd { return nil }

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle       lipgloss.Style = lipgloss.NewStyle().Border(tabBorder).Padding(0, 2)
	activeTabStyle lipgloss.Style = tabStyle.Copy().Bold(true).Border(activeTabBorder, true)

	tabGapStyle lipgloss.Style = tabStyle.Copy().BorderTop(false).BorderLeft(false).BorderRight(false)
)

func (t tui) header() string {
	var todoTab, doneTab string

	if t.view == todo {
		todoTab = activeTabStyle.Render(string(todo))
	} else {
		todoTab = tabStyle.Render(string(todo))
	}

	if t.view == done {
		doneTab = activeTabStyle.Render(string(done))
	} else {
		doneTab = tabStyle.Render(string(done))
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, todoTab, doneTab)

	gap := tabGapStyle.Render(strings.Repeat(" ", max(0, t.w-lipgloss.Width(tabs)-2)))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, gap) + "\n\n"

}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func (t tui) View() string {
	selected := lipgloss.NewStyle().Bold(true)
	if len(t.renderSelection) == 0 { // init population
		t.populateRenderSelection()
	}

	ret := t.header() // todo: stringbuilder
	for i, item := range t.renderSelection {

		if i == int(t.selection) {
			ret += "> "
			ret += selected.Render(item.String())
		} else {
			ret += "  "
			ret += item.String()
		}
		ret += "\n"
	}
	return ret
}

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
