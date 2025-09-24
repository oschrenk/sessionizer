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
