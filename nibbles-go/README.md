# Nibbles Go

A retro grid-based snake game inspired by classic Nibbles/Snake, built with Go and Ebitengine.

## Requirements

- Go 1.24 or newer
- Linux desktop with OpenGL/audio support

## Run

```sh
go run .
```

## Build

```sh
go build -o nibbles-go .
./nibbles-go
```

## Controls

- Arrow keys or WASD: move
- Enter or Space: start
- P or Escape: pause/resume
- R: restart
- M: return to menu after game over

## Rules

- Eat apples to increase score and snake length.
- Eating 15 apples clears the level.
- Level 1 starts slow with no obstacles.
- Level 2 adds border walls.
- Level 3 adds static blocks.
- Level 4 adds cross-shaped walls.
- Level 5 adds maze sections.
- Level 6 and later add moving hazards and faster speed.
- Level 7 and later may use disappearing apples.
- Colliding with walls, obstacles, hazards, or the snake body causes game over.

## Configuration

Gameplay constants are defined in `internal/game/game.go`:

- `GridWidth`, `GridHeight`
- `CellSize`
- `ApplesPerLevel`
- `BaseTPS`, `SpeedPerLevel`, `MaxTPS`

## High Score

High score is saved locally as JSON under the user config directory:

```text
~/.config/nibbles-go/highscore.json
```

## Assets

No copyrighted assets are fetched or bundled. Graphics are drawn procedurally, and sound effects are generated placeholder tones at runtime. The `assets/` directory is included for future CC0 or permissively licensed assets with attribution.
