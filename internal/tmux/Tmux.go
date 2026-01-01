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

// shallowSession represents a session without its windows populated.
// Used internally for parsing tmux output before hydration.
type shallowSession struct {
	Id       string
	Name     string
	Attached bool
	Path     string
}

// shallowWindow represents a window without its panes populated.
// Used internally for parsing tmux output before hydration.
type shallowWindow struct {
	Id            string
	Active        bool
	ActiveClients int
	Name          string
}

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

func parseSession(line string) (shallowSession, bool) {
	result := strings.Split(line, sessionSeparator)
	if len(result) != 4 {
		return shallowSession{}, false
	}
	id := result[0]
	name := result[1]
	attached, _ := strconv.ParseBool(result[2])
	path := result[3]

	session := shallowSession{Id: id, Name: name, Attached: attached, Path: path}
	return session, true
}

func parseWindow(line string) (shallowWindow, bool) {
	result := strings.Split(line, windowSeparator)
	if len(result) != 4 {
		return shallowWindow{}, false
	}
	id := result[0]
	active, _ := strconv.ParseBool(result[1])
	activeClient, _ := strconv.Atoi(result[2])
	name := result[3]

	window := shallowWindow{Id: id, Active: active, ActiveClients: activeClient, Name: name}
	return window, true
}

// hydrateSession converts a shallowSession to a full Session by fetching its windows.
func (s *Server) hydrateSession(shallow shallowSession) (Session, error) {
	windows, err := s.ListWindows(shallow.Id)
	if err != nil {
		return Session{}, err
	}

	return Session{
		Id:       shallow.Id,
		Name:     shallow.Name,
		Attached: shallow.Attached,
		Path:     shallow.Path,
		Windows:  windows,
	}, nil
}

// hydrateWindow converts a shallowWindow to a full Window by fetching its panes.
func (s *Server) hydrateWindow(shallow shallowWindow) (Window, error) {
	panes, err := s.ListPanes(shallow.Id)
	if err != nil {
		return Window{}, err
	}

	return Window{
		Id:            shallow.Id,
		Active:        shallow.Active,
		ActiveClients: shallow.ActiveClients,
		Name:          shallow.Name,
		Panes:         panes,
	}, nil
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

func listSessions(detachedOnly bool, sessionId string) ([]shallowSession, error) {
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
	sessions := []shallowSession{}

	for _, line := range lines {
		session, ok := parseSession(line)
		if !ok {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
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

	return s.hydrateSession(sessions[0])
}

// CurrentWindow returns the currently active window
func (s *Server) CurrentWindow() (Window, error) {
	const windowFormat = "#{window_id}:#{window_active}:#{window_active_clients}:#{window_name}"

	args := []string{
		"display-message",
		"-p",
		windowFormat,
	}

	out, _, err := run(args)
	if err != nil {
		return Window{}, err
	}

	shallow, ok := parseWindow(strings.TrimSpace(out))
	if !ok {
		return Window{}, fmt.Errorf("failed to parse window output: %s", out)
	}

	return s.hydrateWindow(shallow)
}

// Lists all sessions managed by this server.
func (s *Server) ListSessions(detachedOnly bool) ([]Session, error) {
	shallowSessions, err := listSessions(detachedOnly, "")
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(shallowSessions))
	for _, shallow := range shallowSessions {
		session, err := s.hydrateSession(shallow)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Lists all Windows of the targeted session
func (s *Server) ListWindows(sessionId string) ([]Window, error) {
	args := []string{
		"list-windows",
		"-t",
		sessionId,
		"-F",
		"#{window_id}:#{window_active}:#{window_active_clients}:#{window_name}"}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	windows := []Window{}

	for _, line := range lines {
		shallow, ok := parseWindow(line)
		if !ok {
			continue
		}

		window, err := s.hydrateWindow(shallow)
		if err != nil {
			return nil, err
		}

		windows = append(windows, window)
	}

	return windows, nil
}

// Lists all panes in the targeted window
func (*Server) ListPanes(targetWindow string) ([]Pane, error) {
	args := []string{
		"list-panes",
		"-t",
		targetWindow,
		"-F",
		"#{pane_id}:#{pane_index}:#{pane_active}"}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	panes := []Pane{}

	for _, line := range lines {
		result := strings.Split(line, ":")
		if len(result) != 3 {
			continue
		}
		id := result[0]
		index, _ := strconv.Atoi(result[1])
		active, _ := strconv.ParseBool(result[2])

		pane := Pane{Id: id, Index: index, Active: active}

		panes = append(panes, pane)
	}

	return panes, nil
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

// SelectWindow selects (switches to) the specified window.
// The target can be a window ID, window index, or window name.
func (*Server) SelectWindow(targetWindow string) error {
	args := []string{
		"select-window",
		"-t",
		targetWindow,
	}

	_, _, err := run(args)
	return err
}

// SelectPane selects (focuses) the specified pane.
// The target can be a pane ID, or a pane index.
func (*Server) SelectPane(targetPane string) error {
	args := []string{
		"select-pane",
		"-t",
		targetPane,
	}

	_, _, err := run(args)
	return err
}

// SendKeys sends keys/commands to the specified pane.
// Automatically sends Enter (C-m) after the keys.
func (*Server) SendKeys(targetPane string, keys string) error {
	args := []string{
		"send-keys",
		"-t",
		targetPane,
		keys,
		"C-m",
	}

	_, _, err := run(args)
	return err
}

// SplitPane splits the specified pane and returns the new pane ID.
// Direction can be Horizontal (left/right) or Vertical (top/bottom).
func (*Server) SplitPane(targetPane string, direction Direction, startDirectory string) (string, error) {
	args := []string{
		"split-window",
		"-t",
		targetPane,
	}

	if direction == Horizontal {
		args = append(args, "-h")
	}

	args = append(args, "-c", startDirectory, "-P", "-F", "#{pane_id}")

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

// SessionByName retrieves a Session by name.
//
// Returns error if session doesn't exist.
func (s *Server) SessionByName(name string) (*Session, error) {
	const sessionFormat = "#{session_id}:#{session_name}:#{session_attached}:#{session_path}"

	args := []string{
		"display-message",
		"-t",
		name,
		"-p",
		sessionFormat,
	}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	shallow, ok := parseSession(strings.TrimSpace(out))
	if !ok {
		return nil, fmt.Errorf("failed to parse session output: %s", out)
	}

	session, err := s.hydrateSession(shallow)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// Add session
//
// we should guard against session names containing
// - the separator `:`
// - the char `.`
//
// but are problematic, but since we normalize before, we should be fine
func (s *Server) AddSession(name string, path string) (Session, error) {
	name = normalizeName(name)
	const sessionFormat = "#{session_id}:#{session_name}:#{session_attached}:#{session_path}"

	args := []string{
		"new-session",
		"-d",
		"-s",
		name,
		"-c",
		path,
		"-P",
		"-F",
		sessionFormat,
	}
	out, _, err := run(args)
	if err != nil {
		return Session{}, err
	}

	shallow, ok := parseSession(strings.TrimSpace(out))
	if !ok {
		return Session{}, fmt.Errorf("failed to parse session output: %s", out)
	}

	return s.hydrateSession(shallow)
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

func switchSession(sessionName string) error {
	args := []string{
		"switch",
		"-t",
		sessionName,
	}

	_, _, err := run(args)
	if err != nil {
		return err
	}

	return nil
}

func attachSession(sessionName string) error {
	args := []string{
		"attach",
		"-t",
		sessionName,
	}

	_, _, err := run(args)
	if err != nil {
		return err
	}

	return nil
}

func switchClient(sessionName string) error {
	args := []string{
		"switch-client",
		"-t",
		sessionName,
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
func (s *Server) CreateOrAttachSession(name string, path string) (Session, error) {
	var session Session
	sessionPtr, err := s.SessionByName(name)
	if err != nil {
		session, err = s.AddSession(name, path)
		if err != nil {
			return Session{}, err
		}
	} else {
		session = *sessionPtr
	}

	var attachErr error
	switch getContext() {
	case Attached:
		attachErr = switchSession(name)
	case Detached:
		attachErr = attachSession(name)
	case Serverless:
		attachErr = switchClient(name)
	}

	if attachErr != nil {
		return Session{}, attachErr
	}

	return session, nil
}
