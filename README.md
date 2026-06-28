# Niblr

Niblr is a small retro Nibbles-inspired snake game written in Go with Ebitengine. It has deterministic levels, local high scores, difficulty modes, generated beep/boop sounds, and no downloaded assets.

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
| High score entry | Letters/numbers | Enter name |
| High score entry | Backspace | Delete character |
| High score entry | Enter | Save score |
| High score entry | Esc | Cancel |

## Difficulty

- Normal: baseline speed.
- Hard: 2x speed.
- Insane: 4x speed.

Each apple increases speed during the current level. Speed resets when a new level starts.

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

## Credits

- Built with Go and Ebitengine.
- Graphics and sounds are generated procedurally.

## License

Niblr is released under the MIT License. See `LICENSE`.
