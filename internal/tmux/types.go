package tmux

// Server represents a tmux server instance and provides methods to interact with it.
type Server struct {
}

// Session represents a tmux session with its name, attachment status, and working directory.
type Session struct {
	Name     string `json:"name"`
	Attached bool   `json:"attached"`
	Path     string `json:"path"`
}

// Window represents a tmux window within a session.
type Window struct {
	Id            string `json:"id"`
	Active        bool   `json:"active"`
	ActiveClients int    `json:"active_clients"`
	Name          string `json:"name"`
}

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
