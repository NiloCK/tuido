package tui

import (
	"fmt"
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

	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, searchBox, gap, helpPrompt)
}

func (t tui) footer() string {
	footStyle := tabStyle.Copy().BorderBottom(false).BorderLeft(false).BorderRight(false)

	item := t.currentSelection()
	fStr := footStyle.Render(fmt.Sprintf("%s:%d", item.File(), item.Line()))

	// [ ] the [enter] key here is not actually bound to anything
	openPrompt := footStyle.Copy().Faint(true).Render("[enter] - inspect item")
	gap := footStyle.Render(strings.Repeat(" ", max(0, t.w-lipgloss.Width(
		lipgloss.JoinHorizontal(lipgloss.Bottom, fStr, openPrompt))-5,
	)))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, fStr, gap, openPrompt)
}

func (t tui) View() string {
	if t.h == 0 {
		return ""
	}
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
		if len(t.renderSelection) == 0 { // init population
			t.populateRenderSelection()
		}

		header := t.header()
		footer := t.footer()

		rows := []string{}

		availableHeight := t.h - (lipgloss.Height(header) + lipgloss.Height(footer))

		body := t.renderVisibleListedItems(availableHeight)

		rows = append(rows, header, body, footer)
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}
}

func (t tui) renderVisibleListedItems(availableHeight int) string {
	// [ ] needs rendering (somewhere - footer? tab?) of page #
	renderedItems := t.renderedItemCollection()

	pages := []string{}
	newPage := ""

	for _, renderedItem := range renderedItems {
		added := ""
		if newPage == "" {
			added = renderedItem // do not vertically stack the empty page w/ the renderedItem
		} else {
			added = lipgloss.JoinVertical(lipgloss.Left,
				newPage,
				renderedItem,
			)
		}
		if lipgloss.Height(added) <= availableHeight {
			newPage = added
		} else {
			pages = append(pages, newPage)
			newPage = renderedItem
		}
	}
	if len(pages) == 0 || len(newPage) != 0 {
		pages = append(pages, newPage)
	}

	renderedList := pages[t.selection/availableHeight]

	gap := availableHeight - lipgloss.Height(renderedList)
	for i := 0; i < gap; i++ {
		renderedList += "\n"
	}

	return lipgloss.NewStyle().MaxHeight(availableHeight).Render(
		renderedList,
	)
}

func (t tui) renderedItemCollection() []string {
	// [ ]: `selected` style does not apply past the first tag
	selected := lipgloss.NewStyle().Bold(true)

	renderedItems := []string{}

	for i, item := range t.renderSelection {
		renderedItem := ""
		if i == int(t.selection) {
			renderedItem += "> "
			renderedItem += selected.Render(t.renderTuido(*item))
		} else {
			renderedItem += "  "
			renderedItem += t.renderTuido(*item)
		}
		renderedItems = append(renderedItems, renderedItem)
	}
	return renderedItems
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
