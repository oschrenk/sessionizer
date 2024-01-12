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

[projects]
base_dir = "$HOME/Projects" # required
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
```

**List all detached sessions**

Helpful if you want to list alternative sessions eg. in a tmux status bar

```
sessionizer sessions --detached-only
```

## TODO

- [ ] prevent user giving the default session a name containing `.` or `:`
