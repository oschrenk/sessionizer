package tmux

import (
	"os"
	"strings"

	"github.com/oschrenk/sessionizer/shell"
)

type Server struct {
}

type Session struct {
	Name string
	Path string
}

type TmuxContext int64

const (
	Attached TmuxContext = iota
	Detached
	Serverless
)

const sessionSeparator = ":"
const dot = "."
const dash = "-"
const space = " "

func normalizeName(name string) string {
	name = strings.ReplaceAll(name, sessionSeparator, dash)
	name = strings.ReplaceAll(name, space, dash)
	name = strings.ReplaceAll(name, dot, dash)
	return strings.ToLower(name)
}

// "#{session_name}:#{session_path}"}
func sessions(stdout string) ([]Session, error) {
	lines := strings.Split(stdout, "\n")
	sessions := []Session{}

	for _, line := range lines {
		result := strings.Split(line, sessionSeparator)
		if len(result) != 2 {
			continue
		}
		name := result[0]
		path := result[1]

		sessions = append(sessions, Session{Name: name, Path: path})
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
		"#{session_name}:#{session_path}"}
	if detachedOnly {
		args = append(args, []string{"-f", "#{==:#{session_attached},0}"}...)
	}

	out, _, err := run(args)
	if err != nil {
		return nil, err
	}

	return sessions(out)
}

func hasSession(name string) bool {
	args := []string{
		"has-session",
		"-t",
		name,
	}

	_, _, err := run(args)
	return err == nil
}

func addSession(name string, path string) error {
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

// Lists all sessions managed by this server.
func (*Server) ListSessions(detachedOnly bool) ([]Session, error) {
	return listSessions(detachedOnly)
}

func (*Server) HasSession(name string) bool {
	return hasSession(name)
}

// Add session
//
// we should guard against session names containing
// - the separator `:`
// - the char `.`
//
// but are problematic, but since we normalize before, we should be fine
func (*Server) AddSession(name string, path string) error {
	return addSession(name, path)
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

// Attach session
func (*Server) CreateOrAttachSession(name string, path string) error {
	name = normalizeName(name)
	if !hasSession(name) {
		err := addSession(name, path)
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
