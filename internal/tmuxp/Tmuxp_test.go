package tmuxp

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSimple(t *testing.T) {
	layout := Simple("test-window", "/home/user")

	if len(layout.Windows) != 1 {
		t.Errorf("Expected 1 window, got %d", len(layout.Windows))
	}

	window := layout.Windows[0]
	if window.Name != "test-window" {
		t.Errorf("Expected window name 'test-window', got '%s'", window.Name)
	}

	if window.Layout != MainVertical {
		t.Errorf("Expected layout MainVertical, got %s", window.Layout)
	}

	if window.StartDirectory != "/home/user" {
		t.Errorf("Expected start directory '/home/user', got '%s'", window.StartDirectory)
	}

	if len(window.Panes) != 1 {
		t.Errorf("Expected 1 pane, got %d", len(window.Panes))
	}

	pane := window.Panes[0]
	if !pane.Focus {
		t.Error("Expected pane to have focus")
	}

	if len(pane.ShellCommand) != 0 {
		t.Errorf("Expected empty shell command, got %v", pane.ShellCommand)
	}
}

func TestReadLayoutFromFile(t *testing.T) {
	layout, err := ReadLayoutFromFile("testdata/basic_layout.yaml")
	if err != nil {
		t.Fatalf("Failed to read layout from file: %v", err)
	}

	if len(layout.Windows) != 1 {
		t.Errorf("Expected 1 window, got %d", len(layout.Windows))
	}

	window := layout.Windows[0]
	if window.Name != "test" {
		t.Errorf("Expected window name 'test', got '%s'", window.Name)
	}

	if window.Layout != MainVertical {
		t.Errorf("Expected layout MainVertical, got %s", window.Layout)
	}

	if window.StartDirectory != "/tmp" {
		t.Errorf("Expected start directory '/tmp', got '%s'", window.StartDirectory)
	}

	if len(window.Panes) != 1 {
		t.Errorf("Expected 1 pane, got %d", len(window.Panes))
	}

	pane := window.Panes[0]
	if !pane.Focus {
		t.Error("Expected pane to have focus")
	}

	expectedCmd := []string{"echo", "hello"}
	if len(pane.ShellCommand) != len(expectedCmd) {
		t.Errorf("Expected shell command %v, got %v", expectedCmd, pane.ShellCommand)
	}
	for i, cmd := range expectedCmd {
		if pane.ShellCommand[i] != cmd {
			t.Errorf("Expected shell command[%d] '%s', got '%s'", i, cmd, pane.ShellCommand[i])
		}
	}
}

func TestReadLayoutFromFileNotFound(t *testing.T) {
	_, err := ReadLayoutFromFile("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestReadMultiWindowLayout(t *testing.T) {
	layout, err := ReadLayoutFromFile("testdata/multi_window_layout.yaml")
	if err != nil {
		t.Fatalf("Failed to read layout from file: %v", err)
	}

	if len(layout.Windows) != 2 {
		t.Fatalf("Expected 2 windows, got %d", len(layout.Windows))
	}

	// Check first window
	devWindow := layout.Windows[0]
	if devWindow.Name != "dev" {
		t.Errorf("Expected window name 'dev', got '%s'", devWindow.Name)
	}
	if devWindow.Layout != MainVertical {
		t.Errorf("Expected layout MainVertical, got %s", devWindow.Layout)
	}
	if len(devWindow.Panes) != 4 {
		t.Errorf("Expected 4 panes in dev window, got %d", len(devWindow.Panes))
	}

	// Check that pane 3 has focus
	focusCount := 0
	for i, pane := range devWindow.Panes {
		if pane.Focus {
			focusCount++
			if i != 2 {
				t.Errorf("Expected pane 3 (index 2) to have focus, got pane %d", i+1)
			}
		}
	}
	if focusCount != 1 {
		t.Errorf("Expected exactly 1 focused pane, got %d", focusCount)
	}

	// Check second window
	logsWindow := layout.Windows[1]
	if logsWindow.Name != "logs" {
		t.Errorf("Expected window name 'logs', got '%s'", logsWindow.Name)
	}
	if logsWindow.Layout != EvenHorizontal {
		t.Errorf("Expected layout EvenHorizontal, got %s", logsWindow.Layout)
	}
	if logsWindow.StartDirectory != "/var/log" {
		t.Errorf("Expected start directory '/var/log', got '%s'", logsWindow.StartDirectory)
	}
	if len(logsWindow.Panes) != 1 {
		t.Errorf("Expected 1 pane in logs window, got %d", len(logsWindow.Panes))
	}
}

func TestYAMLMarshaling(t *testing.T) {
	layout := Simple("marshal-test", "/test/path")

	data, err := yaml.Marshal(layout)
	if err != nil {
		t.Fatalf("Failed to marshal layout: %v", err)
	}

	var unmarshaled Layout
	err = yaml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal layout: %v", err)
	}

	// Compare the original and unmarshaled layouts
	if len(unmarshaled.Windows) != len(layout.Windows) {
		t.Errorf("Expected %d windows, got %d", len(layout.Windows), len(unmarshaled.Windows))
	}

	if len(unmarshaled.Windows) > 0 {
		origWindow := layout.Windows[0]
		unmarshaledWindow := unmarshaled.Windows[0]

		if origWindow.Name != unmarshaledWindow.Name {
			t.Errorf("Expected window name '%s', got '%s'", origWindow.Name, unmarshaledWindow.Name)
		}

		if origWindow.Layout != unmarshaledWindow.Layout {
			t.Errorf("Expected layout %s, got %s", origWindow.Layout, unmarshaledWindow.Layout)
		}
	}
}
