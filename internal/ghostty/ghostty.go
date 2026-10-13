package ghostty

import (
	"fmt"
	"strings"

	"github.com/oschrenk/sessionizer/internal/shell"
)

// newWindowScript reads the path and the title from argv, so neither is ever
// spliced into the script text. `with title` needs the patched Ghostty fork.
const newWindowScript = `on run argv
  tell application "Ghostty"
    set cfg to new surface configuration
    set initial working directory of cfg to item 1 of argv
    new window with configuration cfg with title (item 2 of argv)
  end tell
end run`

// NewWindow opens a Ghostty window in path, titled title.
func NewWindow(path string, title string) error {
	_, stderr, err := shell.Run("osascript", []string{"-e", newWindowScript, path, title})
	if err != nil {
		return fmt.Errorf("ghostty: %w: %s", err, strings.TrimSpace(stderr))
	}
	return nil
}
