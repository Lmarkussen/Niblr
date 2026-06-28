# Niblr

Niblr is a retro Nibbles-inspired snake game written in Go with Ebitengine.

## Build

```sh
make build
./niblr
```

## Run Without Building

```sh
go run .
```

or:

```sh
make run
```

## Test

```sh
go test ./...
```

or:

```sh
make test
```

## Controls

- Menu Up/Down or 1-3: select Normal, Hard, or Insane
- Menu Space or Enter: start
- Menu H: high scores
- Arrow keys or WASD: move
- P or Esc: pause/unpause
- M: mute/unmute
- Q: quit to main menu
- Space: continue after losing a life or clearing a level
- R: return to main menu after game over
- High score entry: letters/numbers, Backspace, Enter to save, Esc to cancel

## Gameplay

Eat apples to increase your score, grow the snake, and raise the current speed. Each level is cleared after 15 apples, then waits for Space before the next level starts. Speed resets at the start of each level, and later levels add simple static obstacles.

You start with 4 lives. Avoid walls, obstacles, and your own body. Losing a life resets the current level apple count and waits for Space before respawning. Losing all lives ends the game.

## Local Data

Niblr stores settings and high scores as local JSON files. On Linux this uses the XDG config directory:

```text
~/.config/niblr/settings.json
~/.config/niblr/scores.json
```

If the operating system does not provide a config directory, Niblr falls back to the current working directory.
