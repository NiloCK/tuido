package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/nilock/tuido/utils"
)

// upgradeState represents the current state of the upgrade process
type upgradeState int

const (
	upgradePrompt upgradeState = iota
	upgradeDownloading
	upgradeInstalling
	upgradeSuccess
	upgradeError
	upgradeManualInstructions
)

// upgradeModel contains the state for upgrade mode
type upgradeModel struct {
	state           upgradeState
	currentVersion  string
	targetVersion   string
	progress        float64
	downloaded      int64
	total           int64
	errorMessage    string
	manualCommands  []string
	cancelContext   context.Context
	cancelFunc      context.CancelFunc
}

// upgradeMsg represents messages for the upgrade process
type upgradeMsg struct {
	msgType string
	data    interface{}
}

// upgradeProgressMsg represents download progress
type upgradeProgressMsg struct {
	downloaded int64
	total      int64
}

// upgradeCompleteMsg represents successful completion
type upgradeCompleteMsg struct {
	success bool
	result  *utils.UpgradeResult
}

// upgradeErrorMsg represents an error during upgrade
type upgradeErrorMsg struct {
	error error
}

// initUpgradeModel initializes a new upgrade model
func (t *tui) initUpgradeModel() {
	ctx, cancel := context.WithCancel(context.Background())
	
	t.upgrade = upgradeModel{
		state:          upgradePrompt,
		currentVersion: utils.Version(),
		targetVersion:  utils.LatestVersion(),
		progress:       0.0,
		downloaded:     0,
		total:          0,
		cancelContext:  ctx,
		cancelFunc:     cancel,
	}
}

// setUpgradeMode transitions to upgrade mode
func (t *tui) setUpgradeMode() tea.Cmd {
	t.mode = upgrade
	t.initUpgradeModel()
	
	// Check if already on latest version
	if t.upgrade.currentVersion == t.upgrade.targetVersion {
		t.upgrade.state = upgradeError
		t.upgrade.errorMessage = "Already running the latest version"
	}
	
	return nil
}

// startUpgrade begins the upgrade process
func (t *tui) startUpgrade() tea.Cmd {
	t.upgrade.state = upgradeDownloading
	return t.performUpgradeCmd()
}

// performUpgradeCmd executes the upgrade in a goroutine
func (t *tui) performUpgradeCmd() tea.Cmd {
	return func() tea.Msg {
		// Simulate some download progress
		time.Sleep(time.Millisecond * 500)
		
		// Create upgrade configuration
		config := utils.DefaultUpgradeConfig()
		config.Context = t.upgrade.cancelContext
		
		// Perform the upgrade
		result := utils.PerformUpgrade(config)
		
		return upgradeCompleteMsg{
			success: result.Success,
			result:  result,
		}
	}
}

// updateUpgradeModel handles updates for upgrade mode
func (t *tui) updateUpgradeModel(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch t.upgrade.state {
		case upgradePrompt:
			switch msg.String() {
			case "y", "Y", "enter":
				return t.startUpgrade()
			case "n", "N", "esc", "q":
				t.mode = navigation
				return nil
			}
		case upgradeDownloading:
			switch msg.String() {
			case "esc", "q":
				t.upgrade.cancelFunc()
				t.mode = navigation
				return nil
			}
		case upgradeSuccess, upgradeError, upgradeManualInstructions:
			switch msg.String() {
			case "r", "R":
				if t.upgrade.state == upgradeSuccess {
					return tea.Quit // Restart the application
				}
			case "enter", "esc", "q":
				t.mode = navigation
				return nil
			}
		}

	case upgradeProgressMsg:
		t.upgrade.downloaded = msg.downloaded
		t.upgrade.total = msg.total
		if msg.total > 0 {
			t.upgrade.progress = float64(msg.downloaded) / float64(msg.total)
		}

	case upgradeMsg:
		switch msg.msgType {
		case "installing":
			t.upgrade.state = upgradeInstalling
		}

	case upgradeCompleteMsg:
		if msg.success {
			t.upgrade.state = upgradeSuccess
		} else {
			t.upgrade.state = upgradeError
			if msg.result != nil && msg.result.Error != nil {
				t.upgrade.errorMessage = msg.result.Error.Error()
				
				// Check if it's a busy executable error requiring manual steps
				if utils.IsBusyExecutableError(msg.result.Error) {
					t.upgrade.state = upgradeManualInstructions
					currentPath, newPath, _ := utils.GetBusyExecutableInfo(msg.result.Error)
					t.upgrade.manualCommands = []string{
						fmt.Sprintf("mv %s %s", newPath, currentPath),
						"# Then restart tuido",
					}
				}
			}
		}

	case upgradeErrorMsg:
		t.upgrade.state = upgradeError
		t.upgrade.errorMessage = msg.error.Error()
	}

	return nil
}

// renderUpgradeView renders the upgrade mode view
func (t *tui) renderUpgradeView() string {
	switch t.upgrade.state {
	case upgradePrompt:
		return t.renderUpgradePrompt()
	case upgradeDownloading:
		return t.renderUpgradeProgress()
	case upgradeInstalling:
		return t.renderUpgradeInstalling()
	case upgradeSuccess:
		return t.renderUpgradeSuccess()
	case upgradeError:
		return t.renderUpgradeError()
	case upgradeManualInstructions:
		return t.renderUpgradeManualInstructions()
	default:
		return "Unknown upgrade state"
	}
}

// renderUpgradePrompt renders the initial upgrade confirmation prompt
func (t *tui) renderUpgradePrompt() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#00ff00")).
		Render("Upgrade Available")

	currentText := fmt.Sprintf("Current version: %s", t.upgrade.currentVersion)
	targetText := fmt.Sprintf("New version:     %s", t.upgrade.targetVersion)

	versionInfo := lg.NewStyle().
		Margin(1, 0).
		Render(currentText + "\n" + targetText)

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
		Margin(2).
		Width(t.w / 2).
		Render(content)
}

// renderUpgradeProgress renders the download progress
func (t *tui) renderUpgradeProgress() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#0088ff")).
		Render("Downloading Update...")

	versionText := fmt.Sprintf("Upgrading to %s", t.upgrade.targetVersion)

	// Create progress bar
	progressBar := t.renderProgressBar(t.upgrade.progress, 40)
	
	// Format download info
	var sizeInfo string
	if t.upgrade.total > 0 {
		downloaded := formatBytes(t.upgrade.downloaded)
		total := formatBytes(t.upgrade.total)
		percentage := int(t.upgrade.progress * 100)
		sizeInfo = fmt.Sprintf("%s / %s (%d%%)", downloaded, total, percentage)
	} else {
		sizeInfo = "Downloading..."
	}

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
		Margin(2).
		Width(t.w / 2).
		Render(content)
}

// renderUpgradeInstalling renders the installation phase
func (t *tui) renderUpgradeInstalling() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#ffaa00")).
		Render("Installing Update...")

	info := lg.NewStyle().
		Margin(1, 0).
		Render("Replacing executable...")

	spinner := t.renderSpinner()

	content := lg.JoinVertical(lg.Left, title, info, spinner)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(2).
		Width(t.w / 2).
		Render(content)
}

// renderUpgradeSuccess renders the successful upgrade completion
func (t *tui) renderUpgradeSuccess() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#00ff00")).
		Render("✓ Upgrade Successful!")

	info := lg.NewStyle().
		Margin(1, 0).
		Render(fmt.Sprintf("Successfully upgraded to %s", t.upgrade.targetVersion))

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
		Margin(2).
		Width(t.w / 2).
		Render(content)
}

// renderUpgradeError renders upgrade error information
func (t *tui) renderUpgradeError() string {
	title := lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("#ff0000")).
		Render("✗ Upgrade Failed")

	errorInfo := lg.NewStyle().
		Margin(1, 0).
		Foreground(lg.Color("#ff6666")).
		Render("Error: " + t.upgrade.errorMessage)

	suggestion := lg.NewStyle().
		Margin(1, 0).
		Render("You can try upgrading manually by visiting:\n" + utils.ReleaseURL)

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[enter] Continue  [esc] Return to navigation")

	content := lg.JoinVertical(lg.Left, title, errorInfo, suggestion, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(2).
		Width(t.w / 2).
		Render(content)
}

// renderUpgradeManualInstructions renders manual upgrade instructions
func (t *tui) renderUpgradeManualInstructions() string {
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
		Render(strings.Join(t.upgrade.manualCommands, "\n"))

	controls := lg.NewStyle().
		Faint(true).
		Margin(1, 0).
		Render("[enter] Continue  [esc] Return to navigation")

	content := lg.JoinVertical(lg.Left, title, info, commands, controls)

	return lg.NewStyle().
		Align(lg.Left).
		Margin(2).
		Width(t.w / 2).
		Render(content)
}

// renderProgressBar creates a visual progress bar
func (t *tui) renderProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return fmt.Sprintf("[%s]", bar)
}

// renderSpinner creates a simple spinner animation
func (t *tui) renderSpinner() string {
	// Simple spinner using time
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frame := int(time.Now().UnixMilli()/100) % len(frames)
	return frames[frame]
}

// formatBytes formats byte counts into human readable format
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