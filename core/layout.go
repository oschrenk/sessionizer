package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/oschrenk/sessionizer/internal/tmuxp"
)

// ApplyLayout applies a tmuxp layout configuration
//
// MVP: only 1st window and two panes
func ApplyLayout(server *tmux.Server, initialSession tmux.Session, layout tmuxp.Layout) error {

	// Get initial window and pane
	firstLayoutWindow := layout.Windows[0]
	initialWindow := initialSession.Windows[0]
	initialPaneId := initialWindow.Panes[0].Id

	// Rename window if name is specified in layout
	if firstLayoutWindow.Name != "" {
		if err := server.RenameWindow(initialWindow.Id, firstLayoutWindow.Name); err != nil {
			return fmt.Errorf("rename window: %w", err)
		}
	}

	// HACK: Fish shell (and potentially other shells) send terminal capability
	// queries (like ^[[?997;1n for bracketed paste mode) during initialization.
	// These escape sequences can appear in the terminal if we send commands too
	// quickly. We wait for the shell to finish initializing, then clear the
	// screen to remove any visible escape sequences before sending actual commands.
	time.Sleep(200 * time.Millisecond)
	server.SendKeys(initialPaneId, "clear")
	time.Sleep(50 * time.Millisecond)

	// Change directory in first pane
	// Use pane's start_directory if set, otherwise window's start_directory
	firstPaneDir := firstLayoutWindow.Panes[0].StartDirectory
	if firstPaneDir == "" {
		firstPaneDir = firstLayoutWindow.StartDirectory
	}
	if firstPaneDir != "" {
		cdCmd := fmt.Sprintf("cd %s", firstPaneDir)
		if err := server.SendKeys(initialPaneId, cdCmd); err != nil {
			return fmt.Errorf("cd to start directory: %w", err)
		}
	}

	// Track pane IDs for focus selection
	paneIds := []string{initialPaneId}
	var focusedPaneId string
	if firstLayoutWindow.Panes[0].Focus {
		focusedPaneId = initialPaneId
	}

	// Split window and create additional panes
	if len(firstLayoutWindow.Panes) >= 2 {
		// Use second pane's start_directory if set, otherwise window's start_directory
		secondPaneDir := firstLayoutWindow.Panes[1].StartDirectory
		if secondPaneDir == "" {
			secondPaneDir = firstLayoutWindow.StartDirectory
		}

		secondPaneId, err := server.SplitPane(initialPaneId, tmux.Horizontal, secondPaneDir)
		if err != nil {
			return fmt.Errorf("split pane: %w", err)
		}
		paneIds = append(paneIds, secondPaneId)

		if firstLayoutWindow.Panes[1].Focus {
			focusedPaneId = secondPaneId
		}

		// HACK: Same as above - wait for shell initialization and clear escape sequences
		time.Sleep(200 * time.Millisecond)
		server.SendKeys(secondPaneId, "clear")
		time.Sleep(50 * time.Millisecond)
	}

	// Send shell commands to panes
	for i, pane := range firstLayoutWindow.Panes {
		if i >= len(paneIds) {
			break // MVP: only handle first 2 panes
		}
		if len(pane.ShellCommand) > 0 {
			// Join all command arguments into a single string
			cmd := strings.Join(pane.ShellCommand, " ")
			if err := server.SendKeys(paneIds[i], cmd); err != nil {
				return fmt.Errorf("send keys to pane %d: %w", i, err)
			}
		}
	}

	// Select the focused pane
	if focusedPaneId != "" {
		if err := server.SelectPane(focusedPaneId); err != nil {
			return fmt.Errorf("select pane: %w", err)
		}
	}

	return nil
}
