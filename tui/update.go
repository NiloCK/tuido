package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nilock/tuido/tuido"
)

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// [ ] refactor as separate methods per mode
	if _, ok := msg.(tickMsg); ok {
		t.pomoClock--
		if t.pomoClock == 1 {
			t.mode = navigation
		}
		if t.pomoClock < 0 {
			t.pomoClock = 0
		}
		return t, tick()
	}

	if t.mode == nag {
		mode, complete := t.nag.Update(msg)
		t.mode = mode
		if t.mode == navigation && complete { // [ ] need to switch on some nag content?
			t.createNewItem()
		}
		return t, nil
	}

	if t.mode == help {
		if _, ok := msg.(tea.KeyMsg); ok {
			t.mode = navigation
			return t, nil
		}
	}

	if t.mode == pomo {
		if t.pomoClock > 0 {
			return t, nil // no msg processing other than the timer during a running clock
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			str := msg.String()
			if str == "1" ||
				str == "2" ||
				str == "3" ||
				str == "4" ||
				str == "5" ||
				str == "6" ||
				str == "7" ||
				str == "8" ||
				str == "9" ||
				str == "0" ||
				str == "backspace" {
				var cmd tea.Cmd
				t.pomoEditor, cmd = t.pomoEditor.Update(msg)

				return t, cmd
			}
			if str == "enter" {
				t.startPomo()
			}
			if str == "esc" {
				t.pomoEditor.Reset()
				t.mode = navigation
			}
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
		// navigation
		case "up":
			t.setSelection(t.selection - 1)
		case "k":
			t.setSelection(t.selection - 1)
		case "down":
			t.setSelection(t.selection + 1)
		case "j":
			t.setSelection(t.selection + 1)
		case "pgdown": // [ ] these paging functions are not "accurate" #ui #polish
			t.setSelection(t.selection + (len(t.renderSelection) / (t.h - 6)))
		case "pgup":
			t.setSelection(t.selection - (len(t.renderSelection) / (t.h - 6)))
		case "tab":
			t.tab()
		case "/":
			t.filter.Focus()
		case "p":
			t.setPomoMode()
		case "?":
			t.mode = help
		// editing current selection
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
		case "!":
			current := t.currentSelection()
			t.currentSelection().Escalate()
			t.populateRenderSelection()
			for i, item := range t.renderSelection {
				if current == item {
					t.setSelection(i)
				}
			}
		case "e":
			t.setEditMode()
		case "n":
			t.tryCreateNewItem()
		case "z":
			t.currentSelection().Snooze()
		case "q":
			return t, tea.Quit
		}

	case tea.WindowSizeMsg:
		t.h = msg.Height
		t.w = msg.Width
	}
	return t, nil
}

func (t *tui) tryCreateNewItem() {
	if len(t.renderSelection) >= 5 {
		t.setNag("Too many items on your plate...", len(t.renderSelection)-4, navigation)
	} else {
		t.createNewItem()
	}
}

func (t *tui) createNewItem() {
	newItem := tuido.New(t.config.writeto, -1, "")
	t.items = append([]*tuido.Item{&newItem}, t.items...)
	t.populateRenderSelection()
	t.setSelection(0)
	t.setEditMode()
}
