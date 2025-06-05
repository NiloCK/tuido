package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

// demoState represents the current demo state
type demoState int

const (
	demoIntro demoState = iota
	demoUpgradePrompt
	demoDownloading
	demoInstalling
	demoSuccess
	demoError
	demoManualInstructions
	demoEnd
)

// demoModel contains the demo state
type demoModel struct {
	state     demoState
	progress  float64
	autoAdvance bool
	step      int
}

// demoTickMsg represents automatic progression
type demoTickMsg struct{}

func initialModel() demoModel {
	return demoModel{
		state:       demoIntro,
		progress:    0.0,
		autoAdvance: true,
		step:        0,
	}
}

func (m demoModel) Init() tea.Cmd {
	return tea.Tick(time.Second*2, func(time.Time) tea.Msg {
		return demoTickMsg{}
	})
}

func (m demoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ", "enter":
			// Manual advance
			m = m.nextState()
			return m, m.tickCmd()
		case "a":
			// Toggle auto-advance
			m.autoAdvance = !m.autoAdvance
			return m, m.tickCmd()
		}

	case demoTickMsg:
		if m.autoAdvance {
			m = m.nextState()
		}
		return m, m.tickCmd()

	case tea.WindowSizeMsg:
		// Handle window resize if needed
	}

	return m, nil
}

func (m demoModel) nextState() demoModel {
	switch m.state {
	case demoIntro:
		m.state = demoUpgradePrompt
	case demoUpgradePrompt:
		m.state = demoDownloading
		m.progress = 0.0
	case demoDownloading:
		m.progress += 0.33
		if m.progress >= 1.0 {
			m.state = demoInstalling
		}
	case demoInstalling:
		m.state = demoSuccess
	case demoSuccess:
		m.state = demoError
	case demoError:
		m.state = demoManualInstructions
	case demoManualInstructions:
		m.state = demoEnd
	case demoEnd:
		m.state = demoIntro // Loop back
	}
	return m
}

func (m demoModel) tickCmd() tea.Cmd {
	var delay time.Duration
	switch m.state {
	case demoDownloading:
		delay = time.Millisecond * 800
	case demoInstalling:
		delay = time.Second * 1
	default:
		delay = time.Second * 3
	}

	return tea.Tick(delay, func(time.Time) tea.Msg {
		return demoTickMsg{}
	})
}

func (m demoModel) View() string {
	switch m.state {
	case demoIntro:
		return m.renderIntro()
	case demoUpgradePrompt:
		return m.renderUpgradePrompt()
	case demoDownloading:
		return m.renderDownloading()
	case demoInstalling:
		return m.renderInstalling()
	case demoSuccess:
		return m.renderSuccess()
	case demoError:
		return m.renderError()
	case demoManualInstructions:
		return m.renderManualInstructions()
	case demoEnd:
		return m.renderEnd()
	}
	return ""
}

func (m demoModel) renderIntro() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#00ff88")).
		Render("🚀 Tuido Upgrade UI Demo")

	description := lg.NewStyle().
		Margin(1, 0).
		Render("This demo showcases the new in-place upgrade functionality.\nThe upgrade system downloads and replaces the tuido binary automatically.")

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render(fmt.Sprintf("[space/enter] Manual advance  [a] Auto-advance: %v  [q] Quit", m.autoAdvance))

	content := lg.JoinVertical(lg.Left, title, description, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Render(content)
}

func (m demoModel) renderUpgradePrompt() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#00ff00")).
		Render("Upgrade Available")

	versionInfo := lg.NewStyle().
		Margin(1, 0).
		Render("Current version: v0.0.15\nNew version:     v0.0.16")

	prompt := lg.NewStyle().
		Margin(1, 0).
		Render("Do you want to upgrade now?")

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[y] Yes  [n] No  [esc] Cancel")

	content := lg.JoinVertical(lg.Left, title, versionInfo, prompt, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(50).
		Render(content)
}

func (m demoModel) renderDownloading() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#0088ff")).
		Render("Downloading Update...")

	versionText := "Upgrading to v0.0.16"

	progressBar := m.renderProgressBar(m.progress, 40)
	
	downloaded := int64(float64(3355443) * m.progress)
	total := int64(3355443)
	percentage := int(m.progress * 100)
	sizeInfo := fmt.Sprintf("%s / %s (%d%%)", formatBytes(downloaded), formatBytes(total), percentage)

	progressInfo := lg.NewStyle().
		Margin(1, 0).
		Render(progressBar + "\n" + sizeInfo)

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[esc] Cancel download")

	content := lg.JoinVertical(lg.Left, title, versionText, progressInfo, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(50).
		Render(content)
}

func (m demoModel) renderInstalling() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#ffaa00")).
		Render("Installing Update...")

	info := lg.NewStyle().
		Margin(1, 0).
		Render("Replacing executable...")

	spinner := m.renderSpinner()

	content := lg.JoinVertical(lg.Left, title, info, spinner)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(50).
		Render(content)
}

func (m demoModel) renderSuccess() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#00ff00")).
		Render("✓ Upgrade Successful!")

	info := lg.NewStyle().
		Margin(1, 0).
		Render("Successfully upgraded to v0.0.16")

	restartPrompt := lg.NewStyle().
		Margin(1, 0).
		Render("The upgrade is complete. Restart is recommended.")

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[r] Restart tuido  [enter] Continue with current session")

	content := lg.JoinVertical(lg.Left, title, info, restartPrompt, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(60).
		Render(content)
}

func (m demoModel) renderError() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#ff0000")).
		Render("✗ Upgrade Failed")

	errorInfo := lg.NewStyle().
		Margin(1, 0).
		Foreground(lg.Color("#ff6666")).
		Render("Error: Network timeout during download")

	suggestion := lg.NewStyle().
		Margin(1, 0).
		Render("You can try upgrading manually by visiting:\nhttps://github.com/NiloCK/tuido/releases/latest")

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[enter] Continue  [esc] Return to navigation")

	content := lg.JoinVertical(lg.Left, title, errorInfo, suggestion, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(60).
		Render(content)
}

func (m demoModel) renderManualInstructions() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#ffaa00")).
		Render("Manual Upgrade Required")

	info := lg.NewStyle().
		Margin(1, 0).
		Render("The executable is currently in use and cannot be replaced automatically.\nPlease run these commands after restarting:")

	commands := lg.NewStyle().
		Margin(1, 0).
		Background(lg.Color("#333333")).
		Padding(0, 1).
		Render("mv /tmp/tuido_new.12345 ~/bin/tuido\n# Then restart tuido")

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[enter] Continue  [esc] Return to navigation")

	content := lg.JoinVertical(lg.Left, title, info, commands, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(70).
		Render(content)
}

func (m demoModel) renderEnd() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#88ff00")).
		Render("🎉 Demo Complete!")

	summary := lg.NewStyle().
		Margin(1, 0).
		Render("The upgrade UI provides:\n• Clear version information\n• Real-time download progress\n• Error handling and recovery\n• Manual fallback instructions\n• Seamless user experience")

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("Demo will restart automatically...")

	content := lg.JoinVertical(lg.Left, title, summary, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(3).
		Width(50).
		Render(content)
}

func (m demoModel) renderProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return fmt.Sprintf("[%s]", bar)
}

func (m demoModel) renderSpinner() string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frame := int(time.Now().UnixMilli()/100) % len(frames)
	return frames[frame] + " Installing..."
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

func main() {
	fmt.Println("🚀 Starting Tuido Upgrade UI Demo...")
	fmt.Println("Press 'a' to toggle auto-advance, space/enter for manual control")
	fmt.Println()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error running demo: %v", err)
		os.Exit(1)
	}

	fmt.Println("\n✨ Demo completed! The upgrade UI is ready for production.")
}