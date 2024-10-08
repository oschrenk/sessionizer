# README

Yet another tmux sessionizer

## Config

Requires a config file at `$XDG_CONFIG_HOME/sessionizer/config.toml` (defaulting to `$HOME/.config/sessionizer/config.toml`

For example:

```
[base]
ignore = ["node_modules"]   # optional

[default]
name = "default"            # optional
path = "$HOME/Downloads"    # required

[search]
directories = [
  "$HOME/Projects"
]

entries = [
  "$HOME/.local/share/chezmoi"
]
```

## Usage

**Open a fuzzy search**

Lists all projects (directories with a `.git` sub-directory), and, upon selection starting or switching to that tmux session. It always offers the `default` session.

```
sessionizer search
```

**List all sessions**

```
sessionizer sessions
default
personal/project
```

**List all detached sessions**

Helpful if you want to list alternative sessions eg. in a tmux status bar

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
With

- `active` meaning that the window is currently selected by a session as it's window
- `active_clients` counting the number of clients actively viewing that window.

The distinction here is that a session might be detached. A window would still be `active` but would have one less `active_lients`

**Start default session**

```
sessionizer start
```

## TODO

- [ ] prevent user giving the default session a name containing `.` or `:`
