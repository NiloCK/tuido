package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nilock/tuido/tuido"
)

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if t.mode == help {
		if _, ok := msg.(tea.KeyMsg); ok {
			t.mode = navigation
			return t, nil
		}
	}

	if t.mode == edit {
		if msg, ok := msg.(tea.KeyMsg); ok {
			key := msg.String()
			if key == "esc" {
				t.mode = navigation // abandon changes
			}
			if key == "enter" {
				if txt := t.itemEditor.Value(); txt != "" {
					err := t.currentSelection().SetText(txt)
					if err != nil {
						fmt.Println("error: ", err)
					}
					t.mode = navigation
				}
			}
		}

		var cmd tea.Cmd
		t.itemEditor, cmd = t.itemEditor.Update(msg)
		return t, cmd
	}

	t.populateRenderSelection()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if t.filter.Focused() { // [x] replace this w/ the mode-switch as with edit
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
		case "e":
			return t, t.setEditMode()
		case "n":
			newItem := tuido.New(t.config.writeto, -1, "")
			t.items = append([]*tuido.Item{&newItem}, t.items...)
			t.populateRenderSelection()
			t.setSelection(0)
			t.setEditMode()
			return t, t.setEditMode()
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
