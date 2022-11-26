package tui

import (
	"strings"

	lg "github.com/charmbracelet/lipgloss"
	"github.com/nilock/tuido/tuido"
)

type peekScreen struct {
	item tuido.Item
}

func (p *peekScreen) View(h, w int, footer func() string) string {
	foot := footer()

	availableHeight := h - lg.Height(foot)

	bodyPadding := 2

	peekBody, n := p.item.GetContext(availableHeight)
	peekBody = lg.
		NewStyle().
		Padding(bodyPadding).
		Render(peekBody)

	st := lg.NewStyle().
		Foreground(lg.Color("#a0f0a0"))

	pointer := strings.Repeat("  │\n", n+bodyPadding)
	pointer += ">>│"
	pointer += strings.Repeat("\n  │", lg.Height(peekBody)-lg.Height(pointer))

	pointer = st.Render(pointer)

	bodyWithPointer := lg.JoinHorizontal(lg.Top, pointer, peekBody)

	// pad bottom
	availableHeight = h - (lg.Height(peekBody) + lg.Height(foot))
	for i := 0; i < availableHeight; i++ {
		bodyWithPointer += "\n"
	}

	return lg.JoinVertical(lg.Left, bodyWithPointer, foot)
}
