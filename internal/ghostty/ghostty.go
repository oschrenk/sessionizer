package ghostty

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/oschrenk/sessionizer/internal/shell"
	"github.com/oschrenk/sessionizer/internal/tmuxp"
)

// NewWindow opens a Ghostty window titled title, with one tab per layout window
// and one split per pane. `with title` needs the patched Ghostty fork.
func NewWindow(path string, title string, layout tmuxp.Layout) error {
	script, args, err := buildScript(path, title, layout)
	if err != nil {
		return err
	}
	_, stderr, err := shell.Run("osascript", append([]string{"-e", script}, args...))
	if err != nil {
		return fmt.Errorf("ghostty: %w: %s", err, strings.TrimSpace(stderr))
	}
	return nil
}

// script collects AppleScript lines and the argv they read. Every path, name
// and command goes through arg, so none of them is ever part of the script text.
type script struct {
	lines []string
	args  []string
}

func (s *script) arg(v string) string {
	s.args = append(s.args, v)
	return fmt.Sprintf("(item %d of argv)", len(s.args))
}

func (s *script) add(format string, a ...any) {
	s.lines = append(s.lines, "    "+fmt.Sprintf(format, a...))
}

// config sets cfg up for one pane: its directory and its command.
func (s *script) config(dir string, pane tmuxp.Pane) {
	s.add("set cfg to new surface configuration")
	s.add("set initial working directory of cfg to %s", s.arg(dir))
	if len(pane.ShellCommand) > 0 {
		s.add("set initial input of cfg to %s", s.arg(strings.Join(pane.ShellCommand, " ")+"\n"))
	}
}

func buildScript(path string, title string, layout tmuxp.Layout) (string, []string, error) {
	s := &script{}
	var focus []string

	for w, window := range layout.Windows {
		plan, err := splitPlan(window.Layout, len(window.Panes))
		if err != nil {
			return "", nil, fmt.Errorf("window %d: %w", w, err)
		}

		dir := func(p int) string {
			for _, d := range []string{window.Panes[p].StartDirectory, window.StartDirectory} {
				if d != "" {
					return d
				}
			}
			return path
		}
		term := func(p int) string { return fmt.Sprintf("t%d_%d", w, p) }

		s.config(dir(0), window.Panes[0])
		if w == 0 {
			s.add("set w to new window with configuration cfg with title %s", s.arg(title))
			s.add("set tb to selected tab of w")
		} else {
			s.add("set tb to new tab in w with configuration cfg")
		}
		s.add("set %s to focused terminal of tb", term(0))

		if window.Name != "" {
			s.add("perform action (\"set_tab_title:\" & %s) on %s", s.arg(window.Name), term(0))
		}

		for _, st := range plan {
			s.config(dir(st.pane), window.Panes[st.pane])
			s.add("set %s to split %s direction %s with configuration cfg", term(st.pane), term(st.parent), st.dir)
		}
		if len(plan) > 0 {
			s.add("perform action \"equalize_splits\" on %s", term(0))
		}

		for p, pane := range window.Panes {
			if pane.Focus {
				focus = append(focus, term(p))
			}
		}
	}

	// focus selects the tab of the terminal it focuses, so tab 1 goes last, as
	// tmux selects the first window after applying a layout
	for _, t := range slices.Backward(focus) {
		s.add("focus %s", t)
	}
	s.add("select tab (tab 1 of w)")

	text := "on run argv\n  tell application \"Ghostty\"\n" + strings.Join(s.lines, "\n") + "\n  end tell\nend run"
	return text, s.args, nil
}

// step creates pane by splitting parent in direction dir.
type step struct {
	pane   int
	parent int
	dir    string
}

// splitPlan orders the splits that turn one pane into n panes arranged like
// the tmux layout. A chain of splits in one direction comes out even once
// equalize_splits runs.
func splitPlan(layout tmuxp.LayoutType, n int) ([]step, error) {
	plan := []step{}
	chain := func(from int, to int, dir string) {
		for p := from; p < to; p++ {
			plan = append(plan, step{p, p - 1, dir})
		}
	}

	switch layout {
	case tmuxp.EvenHorizontal, "":
		chain(1, n, "right")
	case tmuxp.EvenVertical:
		chain(1, n, "down")
	case tmuxp.MainVertical:
		if n > 1 {
			plan = append(plan, step{1, 0, "right"})
			chain(2, n, "down")
		}
	case tmuxp.MainHorizontal:
		if n > 1 {
			plan = append(plan, step{1, 0, "down"})
			chain(2, n, "right")
		}
	case tmuxp.Tiled:
		// rows first, so the root splits into full-width rows
		cols := int(math.Ceil(math.Sqrt(float64(n))))
		for r := cols; r < n; r += cols {
			plan = append(plan, step{r, r - cols, "down"})
		}
		for r := 0; r < n; r += cols {
			chain(r+1, min(r+cols, n), "right")
		}
	default:
		return nil, fmt.Errorf("unknown layout %q", layout)
	}
	return plan, nil
}
