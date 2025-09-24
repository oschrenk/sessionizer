package tmuxp

type Layout struct {
	Windows []Window `yaml:"windows"`
}

type Window struct {
	Name           string `yaml:"window_name"`
	Layout         string `yaml:"layout"`
	StartDirectory string `yaml:"start_directory"`
	Panes          []Pane `yaml:"panes"`
}

type Pane struct {
	ShellCommand []string `yaml:"shell_command"`
	Focus        bool     `yaml:"focus"`
}
