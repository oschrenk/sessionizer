package tmux

import (
	"fmt"
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

// normalizeName converts a session name to a tmux-safe
// format y replacing problematic characters
// - (colons, spaces, dots) with dashes and converting to lowercase.
//
// This prevents issues with tmux's session name parsing which uses colon as a separator.
func normalizeName(name string) string {
	name = strings.ReplaceAll(name, sessionSeparator, dash)
	name = strings.ReplaceAll(name, space, dash)
	name = strings.ReplaceAll(name, dot, dash)
	return strings.ToLower(name)
}

func run(args []string) (string, string, error) {
	return shell.Run("tmux", args)
}

func parseSession(line string) (Session, bool) {
	result := strings.Split(line, sessionSeparator)
	if len(result) != 4 {
		return Session{}, false
	}
	id := result[0]
	name := result[1]
	attached, _ := strconv.ParseBool(result[2])
	path := result[3]

	session := Session{Id: id, Name: name, Attached: attached, Path: path}
	return session, true
}

func (*Server) currentSessionId() (string, error) {
	args := []string{
		"display-message",
		"-p",
		"#{session_id}"}

	out, _, err := run(args)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func listSessions(detachedOnly bool, sessionId string) ([]Session, error) {
	const sessionFormat = "#{session_id}:#{session_name}:#{session_attached}:#{session_path}"

	args := []string{
		"list-sessions",
		"-F",
		sessionFormat}

	if detachedOnly {
		args = append(args, []string{"-f", "#{==:#{session_attached},0}"}...)
	}

	if sessionId != "" {
		args = append(args, []string{"-f", fmt.Sprintf("#{==:#{session_id},%s}", sessionId)}...)
	}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	sessions := []Session{}

	for _, line := range lines {
		session, ok := parseSession(line)
		if !ok {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// "#{window_id}:#{window_active}:#{window_active_clients}:#{window_name}"
func parseWindows(stdout string) ([]Window, error) {
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

		window := Window{Id: id, Active: active, ActiveClients: activeClient, Name: name}

		windows = append(windows, window)
	}

	return windows, nil
}

func (s *Server) CurrentSession() (Session, error) {
	currentSessionId, err := s.currentSessionId()
	if err != nil {
		return Session{}, err
	}

	sessions, err := listSessions(false, currentSessionId)
	if err != nil {
		return Session{}, err
	}

	if len(sessions) == 0 {
		return Session{}, fmt.Errorf("no session found with id: %s", currentSessionId)
	}

	return sessions[0], nil
}

// Lists all sessions managed by this server.
func (*Server) ListSessions(detachedOnly bool) ([]Session, error) {
	return listSessions(detachedOnly, "")
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

	return parseWindows(out)
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
	_, err := listSessions(false, "")
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
//
// The session name is normalized (lowercased, special chars replaced with dashes) before use.
//
// If the session doesn't exist, it will be created with the specified path as the starting directory.
//
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
