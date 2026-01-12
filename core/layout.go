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
// MVP: only 1st window, supports multiple panes
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
	for i := 1; i < len(firstLayoutWindow.Panes); i++ {
		// Use pane's start_directory if set, otherwise window's start_directory
		paneDir := firstLayoutWindow.Panes[i].StartDirectory
		if paneDir == "" {
			paneDir = firstLayoutWindow.StartDirectory
		}

		newPaneId, err := server.SplitPane(initialPaneId, tmux.Horizontal, paneDir)
		if err != nil {
			return fmt.Errorf("split pane %d: %w", i, err)
		}
		paneIds = append(paneIds, newPaneId)

		if firstLayoutWindow.Panes[i].Focus {
			focusedPaneId = newPaneId
		}

		// HACK: Same as above - wait for shell initialization and clear escape sequences
		time.Sleep(200 * time.Millisecond)
		server.SendKeys(newPaneId, "clear")
		time.Sleep(50 * time.Millisecond)
	}

	// Apply window layout type if specified
	if firstLayoutWindow.Layout != "" {
		if err := server.SelectLayout(initialWindow.Id, string(firstLayoutWindow.Layout)); err != nil {
			return fmt.Errorf("select layout: %w", err)
		}
	}

	// Send shell commands to panes
	for i, pane := range firstLayoutWindow.Panes {
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
