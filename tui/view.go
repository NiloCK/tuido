package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nilock/tuido/tuido"
)

var ( // header styles
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

	if t.itemsFilter == todo {
		todoTab = activeTabStyle.Render(string(todo))
	} else {
		todoTab = tabStyle.Render(string(todo))
	}

	if t.itemsFilter == done {
		doneTab = activeTabStyle.Render(string(done))
	} else {
		doneTab = tabStyle.Render(string(done))
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, todoTab, doneTab)
	searchBox := tabGapStyle.Render(t.filter.View())
	helpPrompt := tabGapStyle.Copy().Faint(true).Render("? - help")
	gap := tabGapStyle.Render(strings.Repeat(" ", max(0, t.w-lipgloss.Width(
		lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, searchBox, helpPrompt))-5),
	))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, searchBox, gap, helpPrompt) + "\n\n"
}

func (t tui) View() string {
	switch t.mode {
	case help:
		ret := "\n[press any key to exit help]\n\n"
		ret += "x: mark done\ns: mark obsolete (strikethrough)\na: mark ongoing (at)\n[space]: mark open\n\n"
		ret += "[tab]:cycle between todo and done tabs\n/: filter todos by tag\n?: enter help\n\n"
		ret += "q: quit"

		txt := lipgloss.NewStyle().Width(28).Align(lipgloss.Left).
			Render("\ntuido reads txt, md, and xit files from the working directory and locates xit style todo items, allowing for quick navigation and discovery.\n\nUpdating an item's status in tuido writes the corresponding change to disk.")
		return lipgloss.JoinHorizontal(lipgloss.Top, "  ", ret, "   ", txt)
	default:
		selected := lipgloss.NewStyle().Bold(true)
		if len(t.renderSelection) == 0 { // init population
			t.populateRenderSelection()
		}

		ret := t.header() // todo: stringbuilder
		for i, item := range t.renderSelection {

			if i == int(t.selection) {
				ret += "> "
				ret += selected.Render(t.renderTuido(*item)) // [ ]: `selected` style does not apply past the first tag
			} else {
				ret += "  "
				ret += t.renderTuido(*item)
			}
			ret += "\n"
		}
		return ret
	}
}

// renderTuido applies tagColor to the items tags and returns the text
func (t tui) renderTuido(item tuido.Item) string {
	ret := item.String()
	tags := item.Tags()

	for _, tag := range tags {
		ret = strings.ReplaceAll(ret, "#"+tag, t.tagColors[tag].Render("#"+tag))
	}

	return ret
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
