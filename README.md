# Lunarium

A simple, colourful letter-learning game for kids. An illustration appears on
screen (e.g. an apple) and you press the matching first letter (**A**). Built
with [Ebiten](https://ebitengine.org/), Go's 2D game engine.

The game is fully self-contained — all images and the font are embedded in the
binary, so there's nothing else to install or unpack.

## Download & play

Grab the latest build from the [**Releases**](../../releases) page.

| Platform | File |
|----------|------|
| Linux (x86-64) | `lunarium-linux-amd64.tar.gz` |
| Windows (x86-64) | `lunarium-windows-amd64.zip` |
| macOS (Intel + Apple Silicon) | `lunarium-macos-universal.zip` |

### Linux
```sh
tar -xzf lunarium-linux-amd64.tar.gz
./lunarium
```

### Windows
Unzip and double-click `lunarium.exe`.

### macOS
The binary is **unsigned** (this is a free hobby project with no paid Apple
Developer account), so after downloading, macOS marks it as quarantined and
Gatekeeper refuses to open it. This is expected — clear the quarantine flag
once and run it:

```sh
unzip lunarium-macos-universal.zip
xattr -d com.apple.quarantine ./lunarium
./lunarium
```

Why the extra step? macOS only quarantines files that arrive via a browser or
other download app; the flag has nothing to do with the binary being unsafe.
The `xattr` command simply removes that download flag. The download is a single
**universal** binary, so it runs natively on both Intel and Apple Silicon Macs.

## How to play

- From the menu, press **1** for *Easy* (a curated subset of letters) or **2**
  for *All letters*.
- An illustration appears — press the key for the **first letter** of its name.
- Press and hold **ESC** to toggle fullscreen.

## Controls

| Key | Action |
|-----|--------|
| `1` / `2` | Select mode from the menu |
| `A`–`Z`, `0`–`9` | Answer the current challenge |
| `ESC` (hold) | Toggle fullscreen |

## Build from source

Requires [Go](https://go.dev/dl/) 1.24+.

```sh
make build     # builds ./lunarium for your OS
make run       # build and run
```

On **Linux** (including a `golang` Docker/Podman container), Ebiten needs
OpenGL/X11/ALSA development headers to compile. Install them once:

```sh
make deps      # apt (Debian/Ubuntu) or dnf (Fedora), auto-sudo if not root
```

Or manually:

```sh
# Debian/Ubuntu (the official golang image)
sudo apt-get install libgl1-mesa-dev libxcursor-dev libxi-dev \
  libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
# Fedora
sudo dnf install mesa-libGL-devel libXcursor-devel libXi-devel \
  libXinerama-devel libXrandr-devel libXxf86vm-devel alsa-lib-devel
```

macOS and Windows need no extra system packages.

## Licensing

This project is licensed in three parts:

- **Code** — [MIT](LICENSE).
- **Artwork** (`assets/abcimg/`, `assets/cat/`) — [CC0 1.0](assets/LICENSE)
  (public domain). The illustrations were generated with ChatGPT; purely
  AI-generated images generally aren't eligible for copyright, so they're
  dedicated to the public domain rather than licensed under a copyright licence.

### Third-party assets

- **Noto Sans** (`assets/NotoSans-Bold.ttf`) — © The Noto Project Authors,
  licensed under the [SIL Open Font License 1.1](assets/NotoSans-OFL.txt).
