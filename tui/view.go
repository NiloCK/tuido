package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	var right string

	if t.err != nil {
		right = lipgloss.
			NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff2222")).
			Render(t.err.Error())
	} else {

		if t.mode == navigation {
			right = footStyle.Render(t.pagination())
		} else if t.mode == edit {
			right = footStyle.Copy().Faint(true).
				Render("[enter] - Save Changes,  [esc] - Discard Changes")
		}
	}

	spacerWidth := max(0, t.w-lg.Width(lg.JoinHorizontal(lg.Bottom, itemStr, right))-5)
	gap := footStyle.Render(strings.Repeat(" ", spacerWidth))

	return lg.JoinHorizontal(lg.Bottom, itemStr, gap, right)
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
	case nag:
		return t.nag.View()
	case pomo:
		ret := t.renderedItemCollection(t.w)[t.selection] + "\n\n"
		if t.pomoTimeRemaining > 0 {
			if t.pomoTimeRemaining > 60 {
				ret += fmt.Sprint(t.pomoTimeRemaining / 60)
			} else {
				ret += fmt.Sprint(t.pomoTimeRemaining)
			}
		} else {
			ret += t.pomoEditor.View()
		}

		return lg.NewStyle().
			Align(lg.Left).
			Margin(2).
			Width(t.w / 2).
			Render(ret)

	case help:
		controls := "\n[press any key to exit help]\n\n"
		controls += "n: new item\ne: edit item\nz: snooze item\n!: escalate item\n1: relax item\np: begin a pomodoro\n\n"
		controls += "x: mark done\ns: mark obsolete (strikethrough)\na: mark ongoing (at)\n[space]: mark open\n\n"
		controls += "[tab]: cycle between todo and done tabs\n/: filter todos by tag\n?: enter help\n\n"
		controls += "q: quit"

		txt := lg.NewStyle().Width(28).Align(lg.Left).
			Render("\n\n\ntuido reads txt, md, and xit files from the working directory and locates xit style todo items, allowing for quick navigation and discovery.\n\nUpdating an item's status in tuido writes the corresponding change to disk.")
		return lg.JoinHorizontal(lg.Top, "  ", controls, "    ", txt)
	default:
		if len(t.renderSelection) == 0 { // init population
			t.populateRenderSelection()
		}

		header := t.header()
		footer := t.footer()

		rows := []string{}

		availableHeight := t.h - (lg.Height(header) + lg.Height(footer))

		body := t.renderVisibleListedItems(availableHeight, t.w)

		// recalculate footer because pages data was set during body render
		rows = append(rows, header, body, t.footer())
		return lg.JoinVertical(lg.Left, rows...)
	}
}

func (t *tui) renderVisibleListedItems(height, width int) string {
	renderedItems := t.renderedItemCollection(width - 1) // providing a margin

	pages := []string{}

	pageUnderConstruction := ""

	for i, renderedItem := range renderedItems {
		pagePlusNextItem := ""

		if pageUnderConstruction == "" {
			pagePlusNextItem = renderedItem // do not vertically stack the empty page w/ the renderedItem
		} else {
			pagePlusNextItem = lg.JoinVertical(lg.Left,
				pageUnderConstruction,
				renderedItem,
			)
		}

		if lg.Height(pagePlusNextItem) <= height {
			pageUnderConstruction = pagePlusNextItem
		} else {
			pages = append(pages, pageUnderConstruction)
			pageUnderConstruction = renderedItem
		}

		if i == t.selection {
			t.currentPage = len(pages)
		}
	}

	if len(pages) == 0 || len(pageUnderConstruction) != 0 {
		pages = append(pages, pageUnderConstruction)
	}

	t.pages = len(pages)

	renderedList := pages[t.currentPage]

	gap := height - lg.Height(renderedList)
	for i := 0; i < gap; i++ {
		renderedList += "\n"
	}

	return lg.NewStyle().MaxHeight(height).Render(
		renderedList,
	)
}

func (t tui) renderedItemCollection(width int) []string {
	// [ ] `selected` style does not apply past the first tag
	selected := lg.NewStyle().Bold(true)

	renderedItems := []string{}

	for i, item := range t.renderSelection {
		renderedItem := ""
		if i == t.selection {
			cursor := "> "
			if t.mode == edit {
				renderedItem = lg.JoinHorizontal(lg.Top, cursor, selected.Render(t.itemEditor.View()))
			} else {
				renderedItem = lg.JoinHorizontal(lg.Top, cursor, selected.Render(t.renderTuido(*item, width)))
			}

		} else {
			leadingSpace := "  "
			renderedItem = lg.JoinHorizontal(lg.Top, leadingSpace, t.renderTuido(*item, width))
		}
		renderedItems = append(renderedItems, renderedItem)
	}
	return renderedItems
}

// renderTuido applies tagColor to the items tags, splits long items
// over multiple lines, and returns the text
func (t tui) renderTuido(item tuido.Item, width int) string {
	ret := item.String()
	tags := item.Tags()

	for _, tag := range tags {
		ret = strings.ReplaceAll(ret, "#"+tag.String(), t.tagColors[tag.Name()].Render("#"+tag.String()))
	}

	// +2 here because of the leading 'cursor' space
	if len(ret)+2 > width {
		rowsRequired := (len(ret) - 4) / (width - 6) // -6 here instead of 4 because of the cursor spaces
		bodyStyle := lg.NewStyle().Height(rowsRequired)

		ret = lg.JoinHorizontal(lg.Top, bodyStyle.Width(4).Render(ret[:4]), bodyStyle.Width(width-6).Render(ret[4:]))
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
