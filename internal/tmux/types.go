package tmux

// Server represents a tmux server instance and provides methods to interact with it.
type Server struct {
}

// Session represents a tmux session with its name, attachment status, and working directory.
type Session struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Attached bool     `json:"attached"`
	Path     string   `json:"path"`
	Windows  []Window `json:"windows"`
}

// Window represents a tmux window within a session.
type Window struct {
	Id            string `json:"id"`
	Active        bool   `json:"active"`
	ActiveClients int    `json:"active_clients"`
	Name          string `json:"name"`
}

// Pane represents a tmux pane within a window.
type Pane struct {
	Id     string `json:"id"`
	Index  int    `json:"index"`
	Active bool   `json:"active"`
}

// Direction represents the split direction for panes.
type Direction int

const (
	// Horizontal splits the pane horizontally (left/right).
	Horizontal Direction = iota
	// Vertical splits the pane vertically (top/bottom).
	Vertical
)

// TmuxContext represents the current execution context relative to tmux.
type TmuxContext int64

const (
	// Attached indicates the process is running inside a tmux session.
	Attached TmuxContext = iota
	// Detached indicates a tmux server is running but the process is outside any session.
	Detached
	// Serverless indicates no tmux server is currently running.
	Serverless
)
