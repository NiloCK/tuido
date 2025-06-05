package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nilock/tuido/utils"
)

// mockTUI simulates the TUI with upgrade notifications
type mockTUI struct {
	mode         int
	notifs       []string
	upgradeState string
}

func (m mockTUI) Init() tea.Cmd {
	return nil
}

func (m mockTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "?":
			m.mode = 3 // help mode
			return m, nil
		case "u":
			if m.mode == 3 && len(m.notifs) > 0 { // help mode with notifications
				m.upgradeState = "upgrade initiated"
				return m, nil
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m mockTUI) View() string {
	switch m.mode {
	case 3: // help mode
		helpText := "HELP MODE\n\n"
		helpText += "Available commands:\n"
		helpText += "n: new item\n"
		helpText += "e: edit item\n"
		if len(m.notifs) > 0 {
			helpText += "u: upgrade to latest version\n"
		}
		helpText += "q: quit\n\n"
		
		if len(m.notifs) > 0 {
			helpText += "NOTIFICATIONS:\n"
			for _, notif := range m.notifs {
				helpText += "• " + notif + "\n"
			}
		}
		
		if m.upgradeState != "" {
			helpText += "\n✅ " + m.upgradeState
		}
		
		return helpText
	default:
		return "NAVIGATION MODE\nPress ? for help, q to quit"
	}
}

func main() {
	fmt.Println("🧪 Testing Upgrade Key Binding Activation")
	fmt.Println()

	// Test 1: Check version detection
	fmt.Printf("Current version: %s\n", utils.Version())
	latest := utils.LatestVersion()
	fmt.Printf("Latest version: %s\n", latest)
	
	upgradeAvailable := utils.Version() != latest
	fmt.Printf("Upgrade available: %v\n", upgradeAvailable)
	fmt.Println()

	// Test 2: Simulate TUI with notifications
	fmt.Println("🎯 Key Binding Test:")
	fmt.Println("1. Without notifications - 'u' should be inactive")
	fmt.Println("2. With notifications - 'u' should be active in help mode")
	fmt.Println()

	// Create mock TUI with notifications (simulating version mismatch)
	mock := mockTUI{
		mode:         0,
		notifs:       []string{fmt.Sprintf("New version available: %s", latest)},
		upgradeState: "",
	}

	fmt.Println("Press ? to enter help mode, then u to test upgrade activation")
	fmt.Println("Press q to quit at any time")
	fmt.Println()

	p := tea.NewProgram(mock)
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Key binding test completed!")
	fmt.Println("The 'u' key should only work in help mode when upgrade notifications are present.")
}