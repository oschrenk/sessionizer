package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/oschrenk/sessionizer/internal/tmuxp"
)

// applyWindowLayout configures a single window according to its layout specification
func applyWindowLayout(server *tmux.Server, windowId string, initialPaneId string, layoutWindow tmuxp.Window, sessionPath string) error {
	// Rename window if name is specified in layout
	if layoutWindow.Name != "" {
		if err := server.RenameWindow(windowId, layoutWindow.Name); err != nil {
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
	// Use pane's start_directory if set, otherwise window's start_directory, otherwise session path
	firstPaneDir := layoutWindow.Panes[0].StartDirectory
	if firstPaneDir == "" {
		firstPaneDir = layoutWindow.StartDirectory
	}
	if firstPaneDir == "" {
		firstPaneDir = sessionPath
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
	if layoutWindow.Panes[0].Focus {
		focusedPaneId = initialPaneId
	}

	// Split window and create additional panes
	for i := 1; i < len(layoutWindow.Panes); i++ {
		// Use pane's start_directory if set, otherwise window's start_directory, otherwise session path
		paneDir := layoutWindow.Panes[i].StartDirectory
		if paneDir == "" {
			paneDir = layoutWindow.StartDirectory
		}
		if paneDir == "" {
			paneDir = sessionPath
		}

		newPaneId, err := server.SplitPane(initialPaneId, tmux.Horizontal, paneDir)
		if err != nil {
			return fmt.Errorf("split pane %d: %w", i, err)
		}
		paneIds = append(paneIds, newPaneId)

		if layoutWindow.Panes[i].Focus {
			focusedPaneId = newPaneId
		}

		// HACK: Same as above - wait for shell initialization and clear escape sequences
		time.Sleep(200 * time.Millisecond)
		server.SendKeys(newPaneId, "clear")
		time.Sleep(50 * time.Millisecond)
	}

	// Apply window layout type if specified
	if layoutWindow.Layout != "" {
		if err := server.SelectLayout(windowId, string(layoutWindow.Layout)); err != nil {
			return fmt.Errorf("select layout: %w", err)
		}
	}

	// Send shell commands to panes
	for i, pane := range layoutWindow.Panes {
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

// ApplyLayout applies a tmuxp layout configuration
//
// Supports multiple windows with multiple panes
func ApplyLayout(server *tmux.Server, initialSession tmux.Session, layout tmuxp.Layout) error {
	// Track the first window ID to return focus at the end
	firstWindowId := initialSession.Windows[0].Id

	// Configure each window in the layout
	for i, layoutWindow := range layout.Windows {
		var windowId string
		var initialPaneId string

		if i == 0 {
			// Use the existing initial window for the first layout window
			windowId = initialSession.Windows[0].Id
			initialPaneId = initialSession.Windows[0].Panes[0].Id
		} else {
			// Create a new window for additional layout windows
			// Use window's start_directory if set, otherwise use session path
			windowDir := layoutWindow.StartDirectory
			if windowDir == "" {
				windowDir = initialSession.Path
			}

			newWindowId, err := server.AddWindow(layoutWindow.Name, windowDir)
			if err != nil {
				return fmt.Errorf("create window %d: %w", i, err)
			}
			windowId = newWindowId

			// Get the initial pane ID of the newly created window
			windows, err := server.ListWindows(initialSession.Id)
			if err != nil {
				return fmt.Errorf("list windows: %w", err)
			}
			// Find the window we just created
			for _, w := range windows {
				if w.Id == newWindowId {
					if len(w.Panes) > 0 {
						initialPaneId = w.Panes[0].Id
					}
					break
				}
			}
			if initialPaneId == "" {
				return fmt.Errorf("could not find pane for new window %d", i)
			}
		}

		// Apply the layout configuration to this window
		if err := applyWindowLayout(server, windowId, initialPaneId, layoutWindow, initialSession.Path); err != nil {
			return fmt.Errorf("apply window %d layout: %w", i, err)
		}
	}

	// Return focus to the first window
	if err := server.SelectWindow(firstWindowId); err != nil {
		return fmt.Errorf("select first window: %w", err)
	}

	return nil
}
