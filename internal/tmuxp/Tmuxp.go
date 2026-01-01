package tmuxp

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// expandPath expands ~ to home directory and environment variables
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = strings.Replace(path, "~", home, 1)
		}
	}
	return os.ExpandEnv(path)
}

type LayoutType string

const (
	EvenHorizontal LayoutType = "even-horizontal"
	EvenVertical   LayoutType = "even-vertical"
	MainHorizontal LayoutType = "main-horizontal"
	MainVertical   LayoutType = "main-vertical"
	Tiled          LayoutType = "tiled"
)

type Layout struct {
	Windows []Window `yaml:"windows"`
}

type Window struct {
	Name           string     `yaml:"window_name"`
	Layout         LayoutType `yaml:"layout"`
	StartDirectory string     `yaml:"start_directory"`
	Panes          []Pane     `yaml:"panes"`
}

type Pane struct {
	ShellCommand   []string `yaml:"shell_command"`
	Focus          bool     `yaml:"focus"`
	StartDirectory string   `yaml:"start_directory"`
}

func Simple(name string, path string) Layout {
	return Layout{
		Windows: []Window{
			{
				Name:           name,
				Layout:         MainVertical,
				StartDirectory: path,
				Panes: []Pane{
					{
						ShellCommand: []string{},
						Focus:        true,
					},
				},
			},
		},
	}
}

func ReadLayoutFromFile(filePath string) (*Layout, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var layout Layout
	err = yaml.Unmarshal(data, &layout)
	if err != nil {
		return nil, err
	}

	if len(layout.Windows) == 0 {
		return nil, fmt.Errorf("layout must have at least one window")
	}

	for i, window := range layout.Windows {
		if len(window.Panes) == 0 {
			return nil, fmt.Errorf("window %d must have at least one pane", i)
		}
		// Expand ~ and environment variables in window start directory
		layout.Windows[i].StartDirectory = expandPath(window.StartDirectory)

		// Expand ~ and environment variables in pane start directories
		for j, pane := range window.Panes {
			if pane.StartDirectory != "" {
				layout.Windows[i].Panes[j].StartDirectory = expandPath(pane.StartDirectory)
			}
		}
	}

	return &layout, nil
}
