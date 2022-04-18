package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
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

	items := []item{}
	for _, f := range files {
		items = append(items, getItems(f)...)
	}

	tuido := tuido{items, 0}
	prog := tea.NewProgram(tuido)

	if err := prog.Start(); err != nil {
		panic(err)
	}
}

type status string

const (
	open     status = "open"
	checked  status = "checked"
	ongoing  status = "ongoing"
	obsolete status = "obsolete"
)

type item struct {
	// metadata

	file string
	line int

	// item data

	raw    string
	status status
	due    time.Time
	tags   []string
}

func getItems(file string) []item {
	prefixes := []string{"[ ]", "[@]", "[x]", "[~]", "[?]"}
	items := []item{}

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	line := 1
	for scanner.Scan() {
		for _, prefix := range prefixes {
			if strings.HasPrefix(scanner.Text(), prefix) {
				items = append(items, item{
					file: file,
					line: line,

					raw:    scanner.Text(),
					status: open,       // todo: switch on prefix
					tags:   nil,        // todo
					due:    time.Now(), //todo
				})
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

type tuido struct {
	items     []item
	selection uint
}

func (t tuido) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (t tuido) Init() tea.Cmd { return nil }

func (t tuido) View() string {
	ret := ""
	for i, item := range t.items {
		if i == int(t.selection) {
			ret += "> "
		}
		ret += item.raw
		ret += "\n"
	}
	return ret
}
