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

## Releasing (maintainers)

Releases are built automatically by [`.github/workflows/release.yml`](.github/workflows/release.yml)
when a version tag is pushed. Cut one with the `Makefile` target, passing the
bump level:

```sh
make release patch   # backwards-compatible fix     v1.2.3 -> v1.2.4  (alias: fix)
make release minor   # backwards-compatible feature  v1.2.3 -> v1.3.0  (alias: feature)
make release major   # incompatible change           v1.2.3 -> v2.0.0  (alias: breaking)
```

The normal path is to release **from a feature branch** (requires the
authenticated [GitHub CLI](https://cli.github.com), `gh`). `make release`:

1. checks the working tree is clean (nothing uncommitted);
2. pushes the branch, opens a PR into the default branch (if one isn't open
   already) and merges it — so the tag points at *merged* code;
3. bumps from the latest release tag, tags the merged tip, and pushes the tag
   (this fires the build);
4. returns you to the branch you started on — it never leaves you on, or
   commits directly to, the default branch.

You can also run it while on the default branch itself, as long as your local
branch is in sync with `origin`; it then tags `HEAD` directly (no PR).

GitHub Actions then builds all three platforms and publishes a Release with the
binaries attached.

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
