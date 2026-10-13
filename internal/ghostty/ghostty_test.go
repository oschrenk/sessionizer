package ghostty

import (
	"slices"
	"strings"
	"testing"

	"github.com/oschrenk/sessionizer/internal/tmuxp"
)

func TestSplitPlan(t *testing.T) {
	tests := []struct {
		layout tmuxp.LayoutType
		n      int
		want   []step
	}{
		{tmuxp.EvenHorizontal, 1, []step{}},
		{tmuxp.EvenHorizontal, 3, []step{{1, 0, "right"}, {2, 1, "right"}}},
		{"", 2, []step{{1, 0, "right"}}},
		{tmuxp.EvenVertical, 3, []step{{1, 0, "down"}, {2, 1, "down"}}},
		{tmuxp.MainVertical, 1, []step{}},
		{tmuxp.MainVertical, 4, []step{{1, 0, "right"}, {2, 1, "down"}, {3, 2, "down"}}},
		{tmuxp.MainHorizontal, 3, []step{{1, 0, "down"}, {2, 1, "right"}}},
		// 2x2: row 2 first, then each row filled to the right
		{tmuxp.Tiled, 4, []step{{2, 0, "down"}, {1, 0, "right"}, {3, 2, "right"}}},
		// 3 panes: two on top, one full-width below
		{tmuxp.Tiled, 3, []step{{2, 0, "down"}, {1, 0, "right"}}},
		{tmuxp.Tiled, 5, []step{{3, 0, "down"}, {1, 0, "right"}, {2, 1, "right"}, {4, 3, "right"}}},
	}
	for _, tt := range tests {
		got, err := splitPlan(tt.layout, tt.n)
		if err != nil {
			t.Fatalf("splitPlan(%q, %d) error: %v", tt.layout, tt.n, err)
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("splitPlan(%q, %d) = %v, want %v", tt.layout, tt.n, got, tt.want)
		}
	}

	if _, err := splitPlan("spiral", 2); err == nil {
		t.Error("splitPlan(\"spiral\", 2) = nil error, want unknown layout")
	}
}

func TestBuildScript(t *testing.T) {
	layout, err := tmuxp.ReadLayoutFromFile("../tmuxp/testdata/multi_window_layout.yaml")
	if err != nil {
		t.Fatal(err)
	}
	layout.Windows[0].Panes[0].StartDirectory = "/it's a dir"

	text, args, err := buildScript("/session", "group/repo", *layout)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"set found to every window whose name is (item 1 of argv)",
		"set w to new window with configuration cfg with title (item 1 of argv)",
		"set t0_1 to split t0_0 direction right with configuration cfg",
		"set t0_3 to split t0_2 direction down with configuration cfg",
		"set tb to new tab in w with configuration cfg",
		"focus t0_2",
		"select tab (tab 1 of w)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("script lacks %q:\n%s", want, text)
		}
	}

	// argv carries every value, and the script text carries none of them
	for _, v := range []string{"/it's a dir", "group/repo", "dev", "echo Pane 3\n", "/var/log", "tail -f system.log\n"} {
		if !slices.Contains(args, v) {
			t.Errorf("args lack %q: %q", v, args)
		}
		if strings.Contains(text, strings.TrimSpace(v)) {
			t.Errorf("script text contains %q", v)
		}
	}
}
