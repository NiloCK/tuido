package tui

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	tuido := tui{items, 0}
	prog := tea.NewProgram(tuido)

	if err := prog.Start(); err != nil {
		panic(err)
	}
}

type tui struct {
	items     []tuido.Item
	selection uint
}

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			t.selection--
		case "down":
			t.selection++
		case "q":
			return t, tea.Quit
		}
	}
	return t, nil
}

func (t tui) Init() tea.Cmd { return nil }

func (t tui) View() string {
	ret := ""
	for i, item := range t.items {
		if i == int(t.selection) {
			ret += "> "
		} else {
			ret += "  "
		}
		ret += item.String()
		ret += "\n"
	}
	return ret
}

func getItems(file string) []tuido.Item {
	prefixes := []string{"[ ]", "[@]", "[x]", "[~]", "[?]"}
	items := []tuido.Item{}

	f, err := os.Open(file)
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
