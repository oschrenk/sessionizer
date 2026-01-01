package tmuxp

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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
	ShellCommand []string `yaml:"shell_command"`
	Focus        bool     `yaml:"focus"`
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

	return &layout, nil
}
