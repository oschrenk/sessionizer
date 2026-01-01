# tmux

> tmux is a terminal multiplexer. It lets you switch between several programs in one terminal, detach (they keep running in the background) and reattach them ...

- a Server has no, one or more **Sessions**
- a Session has one or more **Windows**
- a Window has one or more **Panes**

## Working directory

> When you create a session, its first window (and panes) inherit the current shell’s working directory unless you specify another with -c `<dir>`.

## Splitting Panes

The terminology is confusing

```
tmux split-window       # same as -v  → creates pane below (top/bottom)
tmux split-window -v    # vertical panes (top/bottom) → new pane below
tmux split-window -h    # horizontal panes (left/right) → new pane on right
```

