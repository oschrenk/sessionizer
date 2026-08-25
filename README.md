# README

Yet another tmux sessionizer

## Installation

### nix

```bash
nix profile install github:oschrenk/sessionizer
```

Prebuilt binaries come from the `oschrenk` Cachix cache, which the flake offers
as a substituter.

### home-manager

The flake ships a home-manager module that installs sessionizer and generates
`~/.config/sessionizer/config.toml` plus the `layouts/` directory next to it, so
the configuration below is declared in Nix rather than hand-edited.

```nix
{
  inputs.sessionizer.url = "github:oschrenk/sessionizer";
  # avoids pulling in a second tmux build
  inputs.sessionizer.inputs.nixpkgs.follows = "nixpkgs";

  # in your home-manager configuration:
  imports = [ inputs.sessionizer.homeModules.sessionizer ];

  programs.sessionizer = {
    enable = true;

    default = {
      name = "config";
      path = "$HOME/.config";
    };

    search = {
      directories = [ "$HOME/Projects" ];
      entries = [
        "$HOME/.local/share/chezmoi"
        {
          path = "$HOME/Obsidian/memex";
          name = "self/memex";
          layout = "two-columns";
        }
      ];
    };

    layouts.two-columns.windows = [
      {
        layout = "even-horizontal";
        panes = [ { focus = true; } { focus = false; } ];
      }
    ];
  };
}
```

### From source

- installs to `$GOBIN/sessionizer`

```
git clone git@github.com:oschrenk/sessionizer.git
cd sessionizer
task install
```

## Config

Needs a config file at `$XDG_CONFIG_HOME/sessionizer/config.toml` (or `$HOME/.config/sessionizer/config.toml`):

```
[base]
ignore = ["node_modules"]   # optional
socket_name = "primary"     # optional; tmux -L target for all commands (omit for the default server)
tmux_path = "tmux"          # optional; path to the tmux executable, resolved through PATH by default

[default]
name = "default"            # optional; omit to disable the default session
path = "$HOME/Downloads"    # required
layout_path = "$HOME/.config/sessionizer/default.yml"  # optional, layout for the default session

[search]
directories = [
  "$HOME/Projects"
]

entries = [
  "$HOME/.local/share/chezmoi",
  { path = "$HOME/Obsidian/memex", name = "interest/memex" },
]
```

## Layouts

New sessions can start with a preset window/pane layout ([tmuxp](https://tmuxp.git-pull.com/) format). sessionizer picks the first it finds:

1. `.sessionizer.yml` in the session directory
2. `default.layout_path` (for the default session)
3. `layouts/<name>.yml` next to your config (via `layout` on a `search.entries` object)

No layout found? You get a plain single window.

## Usage

**Target a specific tmux server**

By default sessionizer talks to the default tmux server. To drive a separate, named server (e.g. one per monitor), pass `--socket-name` / `-s` — it maps to `tmux -L <name>` and applies to every command:

```
sessionizer sessions --socket-name primary
sessionizer start --socket-name primary
```

The socket name is resolved from, in order: the `--socket-name` flag, the `SESSIONIZER_SOCKET_NAME` environment variable, `base.socket_name` in the config, then the default server.

**Point at a specific tmux binary**

sessionizer shells out to `tmux`, resolved through `PATH`. When `PATH` can't be relied on — under launchd, say, whose default excludes the nix profile directories — give it an absolute path with `--tmux-path`, the `SESSIONIZER_TMUX_PATH` environment variable, or `base.tmux_path` in the config, in that order of precedence.

**Open a fuzzy search**

Fuzzy-find a project (any directory with a `.git`) and start or switch to its tmux session. The default session is also offered when `default.name` is set.

```
sessionizer search
```

**Print selected project path**

Same finder, but prints the selected path to stdout instead of starting a session — handy for shell wrappers (e.g. `cd` to it). Silent if cancelled.

```
sessionizer search --print-path
```

**List all sessions**

```
sessionizer sessions
default
personal/project
```

**List all detached sessions**

Handy for listing alternative sessions, e.g. in a tmux status bar.

```
sessionizer sessions --detached-only
```

**List all sessions as json**

```
sessionizer sessions --json
[
  {
    "name": "default",
    "path": "/Users/person/Downloads",
    "attached": false
  },
  {
    "name": "personal/project",
    "path": "/Users/person/Projects/personal/project",
    "attached": true
  },
  ...
]
```

**List windows of attached session (as json)**

```
sessionizer windows --json
[
  {
    "id": "@10",
    "active": true,
    "active_clients": 1,
    "name": "fish"
  }
]
```
Where:

- `active` — the window is the one currently selected by its session
- `active_clients` — how many clients are actively viewing it

A detached session's window can still be `active`, just with one fewer `active_clients`.

**Start a session**

Start or attach to a session. The name comes from `default.name`, or `-n` to override it:

```
sessionizer start            # uses default.name from config
sessionizer start -n work    # start/attach a session named "work"
```

## TODO

- [ ] prevent user giving the default session a name containing `.` or `:`
