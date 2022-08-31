package tui

import (
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

func NewNag(prompt string, size int, exit mode) nagScreen {
	nagLength := fib(size)
	nagText := ""

	for i := 0; i < nagLength; i++ {
		nagText += string(rune('a' + rand.Intn(26)))
	}

	return nagScreen{prompt, nagText, exit}
}

type nagScreen struct {
	prompt  string
	nagText string
	exit    mode
}

func (n *nagScreen) View() string {

	s := lg.NewStyle().Margin(2)

	prompt := s.Render(n.prompt)
	str := s.Render("type \"" + n.nagText + "\" to continue.")
	footer := s.Faint(true).Render("esc: back to item navigation")

	return lg.JoinVertical(lg.Left, prompt, str, footer)
}

func (n *nagScreen) Update(msg tea.Msg) (mode, bool) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc":
			return navigation, false
		default:
			// process keystroke
			if msg.String() == string(n.nagText[0]) {
				n.nagText = n.nagText[1:]
			}
			if len(n.nagText) == 0 {
				return navigation, true
			}
			return nag, false
		}
	}
	return nag, false
}

func (t *tui) setNag(prompt string, size int, exit mode) {
	t.nag = NewNag(prompt, size, exit)
	t.mode = nag
}

var fibs map[int]int = map[int]int{}

func fib(n int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	if n == 2 {
		return 2
	}

	if known, ok := fibs[n]; ok {
		return known
	}

	newOne := fib(n-1) + fib(n-2)
	fibs[n] = newOne

	return newOne
}
