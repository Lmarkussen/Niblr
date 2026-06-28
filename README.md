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
- R: restart after game over

## Gameplay

Eat apples to increase your score and grow the snake. Each level is cleared after 15 apples. Later levels increase the snake speed and add simple static obstacles.

Avoid walls, obstacles, and your own body.
