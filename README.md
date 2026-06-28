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

- Arrow keys or WASD: move
- P or Esc: pause/unpause
- Space: continue after clearing a level
- R: restart after game over

## Gameplay

Eat apples to increase your score, grow the snake, and raise the current speed. Each level is cleared after 15 apples, then waits for Space before the next level starts. Speed resets at the start of each level, and later levels add simple static obstacles.

You start with 4 lives. Avoid walls, obstacles, and your own body. Losing all lives ends the game.
