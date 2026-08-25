package tmux

import (
	"slices"
	"testing"
)

func TestWithSocket(t *testing.T) {
	// socketName is package-global; reset it after the test.
	t.Cleanup(func() { SetSocket("") })

	tests := []struct {
		name   string
		socket string
		args   []string
		want   []string
	}{
		{
			name:   "no socket leaves args unchanged",
			socket: "",
			args:   []string{"list-sessions", "-F", "x"},
			want:   []string{"list-sessions", "-F", "x"},
		},
		{
			name:   "socket prepends -L <name>",
			socket: "primary",
			args:   []string{"list-sessions"},
			want:   []string{"-L", "primary", "list-sessions"},
		},
		{
			name:   "socket with empty args",
			socket: "primary",
			args:   []string{},
			want:   []string{"-L", "primary"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetSocket(tt.socket)
			got := withSocket(tt.args)
			if !slices.Equal(got, tt.want) {
				t.Errorf("withSocket(%v) with socket %q = %v, want %v", tt.args, tt.socket, got, tt.want)
			}
		})
	}
}

func TestSetBinary(t *testing.T) {
	// tmuxBin is package-global; reset it after the test.
	t.Cleanup(func() { tmuxBin = "tmux" })

	tests := []struct {
		name string
		set  string
		want string
	}{
		{
			name: "empty path keeps the default",
			set:  "",
			want: "tmux",
		},
		{
			name: "absolute path replaces the default",
			set:  "/nix/store/abc/bin/tmux",
			want: "/nix/store/abc/bin/tmux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmuxBin = "tmux"
			SetBinary(tt.set)
			if tmuxBin != tt.want {
				t.Errorf("SetBinary(%q) = %q, want %q", tt.set, tmuxBin, tt.want)
			}
		})
	}
}
