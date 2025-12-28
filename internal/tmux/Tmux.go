package tmux

import (
	"os"
	"strconv"
	"strings"

	"github.com/oschrenk/sessionizer/internal/shell"
)

const sessionSeparator = ":"
const windowSeparator = ":"
const dot = "."
const dash = "-"
const space = " "

// normalizeName converts a session name to a tmux-safe format by replacing
// problematic characters (colons, spaces, dots) with dashes and converting to lowercase.
// This prevents issues with tmux's session name parsing which uses colon as a separator.
func normalizeName(name string) string {
	name = strings.ReplaceAll(name, sessionSeparator, dash)
	name = strings.ReplaceAll(name, space, dash)
	name = strings.ReplaceAll(name, dot, dash)
	return strings.ToLower(name)
}

// "#{session_name}:#{session_attached}:#{session_path}"}
func sessions(stdout string) ([]Session, error) {
	lines := strings.Split(stdout, "\n")
	sessions := []Session{}

	for _, line := range lines {
		result := strings.Split(line, sessionSeparator)
		if len(result) != 3 {
			continue
		}
		name := result[0]
		attached, _ := strconv.ParseBool(result[1])
		path := result[2]

		sessions = append(sessions, Session{Name: name, Attached: attached, Path: path})
	}

	return sessions, nil
}

func run(args []string) (string, string, error) {
	return shell.Run("tmux", args)
}

func listSessions(detachedOnly bool) ([]Session, error) {
	args := []string{
		"list-sessions",
		"-F",
		"#{session_name}:#{session_attached}:#{session_path}"}
	if detachedOnly {
		args = append(args, []string{"-f", "#{==:#{session_attached},0}"}...)
	}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	return sessions(out)
}

// "#{window_id}:#{window_active}:#{window_active_clients}:#{window_name}"
func windows(stdout string) ([]Window, error) {
	lines := strings.Split(stdout, "\n")
	windows := []Window{}

	for _, line := range lines {
		result := strings.Split(line, windowSeparator)
		if len(result) != 4 {
			continue
		}
		id := result[0]
		active, _ := strconv.ParseBool(result[1])
		activeClient, _ := strconv.Atoi(result[2])
		name := result[3]

		windows = append(windows, Window{Id: id, Active: active, ActiveClients: activeClient, Name: name})
	}

	return windows, nil
}

// Lists all sessions managed by this server.
func (*Server) ListSessions(detachedOnly bool) ([]Session, error) {
	return listSessions(detachedOnly)
}

// Lists all Windows of the current sessions
func (*Server) ListWindows() ([]Window, error) {
	args := []string{
		"list-windows",
		"-F",
		"#{window_id}:#{window_active}:#{window_active_clients}:#{window_name}"}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	return windows(out)
}

// Creates a new window with the given name and starting directory
// Returns the unique window ID assigned by tmux
func (*Server) AddWindow(name string, path string) (string, error) {
	args := []string{
		"new-window",
		"-n",
		name,
		"-c",
		path,
		"-P",
		"-F",
		"#{window_id}",
	}
	out, _, err := run(args)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// HasSession checks if a tmux session with the given name exists.
// Returns true if the session exists, false otherwise.
func (*Server) HasSession(name string) bool {
	args := []string{
		"has-session",
		"-t",
		name,
	}

	_, _, err := run(args)
	return err == nil
}

// Add session
//
// we should guard against session names containing
// - the separator `:`
// - the char `.`
//
// but are problematic, but since we normalize before, we should be fine
func (*Server) AddSession(name string, path string) error {
	args := []string{
		"new-session",
		"-d",
		"-s",
		name,
		"-c",
		path,
	}
	_, _, err := run(args)
	if err != nil {
		return err
	}
	return nil
}

func getContext() TmuxContext {
	_, err := listSessions(false)
	if err != nil {
		return Serverless
	}

	// if $TMUX is set we are already inside TMUX
	if os.Getenv("TMUX") != "" {
		return Attached
	} else {
		return Detached
	}
}

func switchSession(name string) error {
	args := []string{
		"switch",
		"-t",
		name,
	}

	_, _, err := run(args)
	if err != nil {
		return err
	}

	return nil
}

func attachSession(name string) error {
	args := []string{
		"attach",
		"-t",
		name,
	}

	_, _, err := run(args)
	if err != nil {
		return err
	}

	return nil
}

func switchClient(name string) error {
	args := []string{
		"switch-client",
		"-t",
		name,
	}

	_, _, err := run(args)
	if err != nil {
		return err
	}

	return nil
}

// CreateOrAttachSession creates a new session or attaches to an existing one with the given name.
// The session name is normalized (lowercased, special chars replaced with dashes) before use.
// If the session doesn't exist, it will be created with the specified path as the starting directory.
// The behavior depends on the current tmux context:
//   - Attached: switches to the session using switch-session
//   - Detached: attaches to the session using attach-session
//   - Serverless: switches the client to the session using switch-client
func (s *Server) CreateOrAttachSession(name string, path string) error {
	name = normalizeName(name)
	if !s.HasSession(name) {
		err := s.AddSession(name, path)
		if err != nil {
			return err
		}
	}

	switch getContext() {
	case Attached:
		return switchSession(name)
	case Detached:
		return attachSession(name)
	case Serverless:
		return switchClient(name)
	}

	// should never be reached
	return nil
}
