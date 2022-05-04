package tui

import (
	"fmt"
	"strings"

	lg "github.com/charmbracelet/lipgloss"
	"github.com/nilock/tuido/tuido"
)

var ( // header styles
	activeTabBorder = lg.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lg.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle       lg.Style = lg.NewStyle().Border(tabBorder).Padding(0, 2)
	activeTabStyle lg.Style = tabStyle.Copy().Bold(true).Border(activeTabBorder, true)

	tabGapStyle lg.Style = tabStyle.Copy().BorderTop(false).BorderLeft(false).BorderRight(false)
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

	tabs := lg.JoinHorizontal(lg.Bottom, todoTab, doneTab)
	searchBox := tabGapStyle.Render(t.filter.View())
	helpPrompt := tabGapStyle.Copy().Faint(true).Render("? - help")
	gap := tabGapStyle.Render(strings.Repeat(" ", max(0, t.w-lg.Width(
		lg.JoinHorizontal(lg.Bottom, tabs, searchBox, helpPrompt))-5),
	))

	return lg.JoinHorizontal(lg.Bottom, tabs, searchBox, gap, helpPrompt)
}

func (t tui) footer() string {
	footStyle := tabStyle.Copy().BorderBottom(false).BorderLeft(false).BorderRight(false)

	itemLoc := t.currentSelection().Location()
	itemStr := footStyle.Render(itemLoc)

	pagination := footStyle.Render(t.pagination())

	spacerWidth := max(0, t.w-lg.Width(lg.JoinHorizontal(lg.Bottom, itemStr, pagination))-5)
	gap := footStyle.Render(strings.Repeat(" ", spacerWidth))

	return lg.JoinHorizontal(lg.Bottom, itemStr, gap, pagination)
}

func (t tui) pagination() string {
	ret := ""
	bold := lg.NewStyle().Bold(true).SetString("●")
	faint := lg.NewStyle().Faint(true).SetString("●")

	if t.pages > 1 {
		if t.pages < 8 {
			for i := 0; i < t.pages; i++ {
				if i == t.currentPage {
					ret += bold.String()
				} else {
					ret += faint.String()
				}
			}
		} else {
			ret = faint.Render(fmt.Sprintf("%d of %d", t.currentPage+1, t.pages))
		}
	}
	return ret
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

		txt := lg.NewStyle().Width(28).Align(lg.Left).
			Render("\ntuido reads txt, md, and xit files from the working directory and locates xit style todo items, allowing for quick navigation and discovery.\n\nUpdating an item's status in tuido writes the corresponding change to disk.")
		return lg.JoinHorizontal(lg.Top, "  ", ret, "   ", txt)
	default:
		if len(t.renderSelection) == 0 { // init population
			t.populateRenderSelection()
		}

		header := t.header()
		footer := t.footer()

		rows := []string{}

		availableHeight := t.h - (lg.Height(header) + lg.Height(footer))

		body := t.renderVisibleListedItems(availableHeight)

		// recalculate footer because pages data was set during body render
		rows = append(rows, header, body, t.footer())
		return lg.JoinVertical(lg.Left, rows...)
	}
}

func (t *tui) renderVisibleListedItems(availableHeight int) string {
	// [ ] needs rendering (somewhere - footer? tab?) of page #
	renderedItems := t.renderedItemCollection()

	pages := []string{}
	newPage := ""

	for _, renderedItem := range renderedItems {
		added := ""
		if newPage == "" {
			added = renderedItem // do not vertically stack the empty page w/ the renderedItem
		} else {
			added = lg.JoinVertical(lg.Left,
				newPage,
				renderedItem,
			)
		}
		if lg.Height(added) <= availableHeight {
			newPage = added
		} else {
			pages = append(pages, newPage)
			newPage = renderedItem
		}
	}
	if len(pages) == 0 || len(newPage) != 0 {
		pages = append(pages, newPage)
	}

	t.pages = len(pages)
	t.currentPage = t.selection / availableHeight

	renderedList := pages[t.currentPage]

	gap := availableHeight - lg.Height(renderedList)
	for i := 0; i < gap; i++ {
		renderedList += "\n"
	}

	return lg.NewStyle().MaxHeight(availableHeight).Render(
		renderedList,
	)
}

func (t tui) renderedItemCollection() []string {
	// [ ]: `selected` style does not apply past the first tag
	selected := lg.NewStyle().Bold(true)

	renderedItems := []string{}

	for i, item := range t.renderSelection {
		renderedItem := ""
		if i == t.selection {
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

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
