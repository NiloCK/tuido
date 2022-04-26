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

	items := []tuido.Item{}
	for _, f := range files {
		items = append(items, getItems(f)...)
	}

	tuido := tui{items, nil, todo, 0}
	prog := tea.NewProgram(tuido)

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
	items           []tuido.Item
	renderSelection []tuido.Item
	view            view
	selection       uint
}

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			t.selection-- // [ ] make a fcn w/ border logic
		case "down":
			t.selection++
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
//
// [ ] reset currentSelection to something in the range of
//     the current renderitems
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
}

func (t tui) Init() tea.Cmd { return nil }

func (t tui) View() string {
	selected := lipgloss.NewStyle().Bold(true)
	t.populateRenderSelection()

	ret := "" // todo: stringbuilder
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

func getItems(file string) []tuido.Item {
	prefixes := []string{"[ ]", "[@]", "[x]", "[~]", "[?]"}
	items := []tuido.Item{}

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
				items = append(items, tuido.New(file, line, scanner.Text()))
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
