# Niblr

Niblr is a small retro Nibbles-inspired snake game written in Go with Ebitengine. It has 30 deterministic levels, local high scores, difficulty modes, generated beep/boop sounds, and no downloaded assets.

## Screenshots

Screenshots will be added for the first tagged release.

## Build And Run

Run from source:

```sh
make run
```

Build a local binary:

```sh
make build
./niblr
```

Run tests:

```sh
make test
```

Create a Linux amd64 release archive:

```sh
make release-linux VERSION=v0.1.0
```

This writes:

```text
dist/niblr
dist/niblr-linux-amd64.tar.gz
```

The archive contains the `niblr` binary, `README.md`, `LICENSE`, and `CHANGELOG.md`.

## Controls

| Context | Input | Action |
| --- | --- | --- |
| Menu | Up/Down or `1`-`3` | Select Normal, Hard, or Insane |
| Menu | Space or Enter | Start game |
| Menu | `H` | View high scores |
| Game | Arrow keys or WASD | Move |
| Game | `P` or Esc | Pause/unpause |
| Game | `M` | Mute/unmute |
| Game | `Q` | Quit to main menu |
| Life lost / level clear | Space | Continue |
| Game over | `R` | Return to main menu |
| Victory | `R` | Restart |
| Victory | Esc | Return to main menu |
| High score entry | Letters/numbers | Enter name |
| High score entry | Backspace | Delete character |
| High score entry | Enter | Save score |
| High score entry | Esc | Cancel |

## Controller Support

Niblr supports common standard-layout controllers such as Xbox, PlayStation, and Switch-style controllers. A controller is optional; keyboard input continues to work normally.

| Controller input | Action |
| --- | --- |
| D-pad | Move / menu selection |
| Left analog stick | Move / menu selection, with deadzone |
| A / Cross / primary button | Confirm, start, continue |
| B / Circle / secondary button | Back/cancel where applicable |
| Start / Menu | Pause/unpause |
| Select / Back | Mute/unmute |

## Difficulty

- Normal: fast baseline speed.
- Hard: 2x Normal speed.
- Insane: slightly faster than Hard and intentionally brutal.

Each apple increases speed during the current level. Speed resets when a new level starts.

## Levels

Niblr has 30 deterministic levels. Levels 1-3 introduce the game gently, levels 4-14 add lanes and partial mazes, levels 15-25 become dense and planning-heavy, and levels 26-30 are late-game Nibbles-style challenge layouts.

## Music And Sound

Niblr uses generated beep/boop sound effects and three generated looping chiptune-style tracks for menu, early-level, and late-level play. No external music is bundled. See `assets/README.md` for asset notes.

## Local Data

Niblr stores settings and high scores as local JSON files. On Linux this uses the XDG config directory:

```text
~/.config/niblr/settings.json
~/.config/niblr/scores.json
```

If the operating system does not provide a config directory, Niblr falls back to the current working directory.

## Development

```sh
make fmt
make test
make build
```

## Developer/testing options

These command-line flags are intended for local testing and bypass the normal start menu only when supplied:

```sh
go run . --level 12 --difficulty hard --lives 6 --mute
```

| Flag | Description |
| --- | --- |
| `--level N` | Start directly on level `N`, from 1 to 30 |
| `--difficulty normal\|hard\|insane` | Start with the selected difficulty |
| `--lives N` | Start with `N` lives |
| `--mute` | Start muted |

Invalid level, difficulty, or lives values fail before the game window opens.

## Credits

- Built with Go and Ebitengine.
- Graphics and sounds are generated procedurally.

## License

Niblr is released under the MIT License. See `LICENSE`.
