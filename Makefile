# Lunarium — build & release
#
#   make            # build ./lunarium for the host OS (default)
#   make run        # build and run
#   make install    # build, then copy the binary to $(PREFIX)/bin (default /usr/local)
#   make uninstall  # remove the installed binary
#   make app        # macOS: bundle bin/Lunarium.app (with icon) — needs sips/iconutil
#   make install-app # macOS: build the bundle and copy it to /Applications
#   make deps       # install Ebiten's Linux build headers (Debian/Fedora)
#   make clean
#   make tidy       # go mod tidy
#
#   make release patch   # backwards-compatible fix   v1.2.3 -> v1.2.4  (alias: fix)
#   make release minor   # backwards-compatible feature v1.2.3 -> v1.3.0 (alias: feature)
#   make release major   # incompatible change          v1.2.3 -> v2.0.0 (alias: breaking)
#
# `release` runs scripts/release.sh: from a feature branch it pushes, opens +
# merges a PR into the default branch, then tags the MERGED tip and pushes the
# tag (which fires .github/workflows/release.yml) — and always returns you to
# the branch you started on, never parking on the default branch. Needs the
# GitHub CLI `gh` (authenticated) when run from a feature branch. All the logic
# lives in the script — the Makefile stays trivial so a bare `make` always
# works, regardless of make version.

BINARY := lunarium

# Install location. Override with `make install PREFIX=~/.local`.
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin

# macOS .app bundle settings.
APP        := Lunarium.app
APPDIR     := bin/$(APP)
ICON_SRC   := assets/cat/head.png
BUNDLE_ID  := com.flocko-motion.lunarium
# CFBundle version strings can't carry git's "v"/"-dirty" cruft; strip to a bare
# number-ish string, falling back to 0 outside a tagged checkout.
APP_VERSION := $(patsubst v%,%,$(firstword $(subst -, ,$(VERSION_STAMP))))

# Stamped into the binary via -ldflags (kept simple: no nested parens, so it
# parses on every make version). Falls back to "dev" outside a git checkout.
VERSION_STAMP := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS       := -X main.version=$(VERSION_STAMP)

.DEFAULT_GOAL := build
.PHONY: build run install uninstall app install-app clean tidy deps release major minor patch breaking feature fix

build:
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY) . || { \
		echo ""; \
		echo ">> build failed. On Linux, Ebiten needs OpenGL/X11 dev headers."; \
		echo ">> Try:  make deps   (then re-run make). See README for details."; \
		exit 1; }

run:
	go run -ldflags "$(LDFLAGS)" .

# Build, then copy the binary onto $PATH. Assets are embedded, so the single
# binary is fully self-contained. Uses sudo only when the target isn't writable
# (e.g. the default /usr/local/bin); for a no-sudo install use PREFIX=~/.local.
# Writability is judged against the nearest existing ancestor of $(BINDIR), so a
# not-yet-created dir under a writable parent doesn't needlessly demand sudo.
install: build
	@sudo=""; \
	d="$(BINDIR)"; while [ ! -e "$$d" ] && [ "$$d" != "/" ] && [ "$$d" != "." ]; do d=$$(dirname "$$d"); done; \
	if [ ! -w "$$d" ] && [ "$$(id -u)" != 0 ]; then \
		if command -v sudo >/dev/null 2>&1; then sudo="sudo"; \
		else echo "Can't write to $(BINDIR) and 'sudo' is unavailable — retry with PREFIX=~/.local"; exit 1; fi; \
	fi; \
	$$sudo mkdir -p "$(BINDIR)" && \
	$$sudo install -m 0755 $(BINARY) "$(BINDIR)/$(BINARY)" && \
	echo ">> installed $(BINARY) to $(BINDIR)/$(BINARY)"

uninstall:
	@sudo=""; \
	if [ -e "$(BINDIR)/$(BINARY)" ] && [ ! -w "$(BINDIR)" ] && [ "$$(id -u)" != 0 ]; then \
		if command -v sudo >/dev/null 2>&1; then sudo="sudo"; fi; \
	fi; \
	$$sudo rm -f "$(BINDIR)/$(BINARY)" && \
	echo ">> removed $(BINDIR)/$(BINARY)"

# Bundle a double-clickable macOS app with a Dock/Finder icon. macOS-only:
# sips + iconutil ship with the OS. The .icns is generated from $(ICON_SRC)
# (512px source is upscaled to fill the 1024px @2x slot).
app: build
	@command -v iconutil >/dev/null 2>&1 && command -v sips >/dev/null 2>&1 || { \
		echo ">> make app is macOS-only (needs sips + iconutil)"; exit 1; }
	@rm -rf $(APPDIR)
	@mkdir -p $(APPDIR)/Contents/MacOS $(APPDIR)/Contents/Resources
	@iconset="$$(mktemp -d)/$(BINARY).iconset"; mkdir -p "$$iconset"; \
	for s in 16 32 128 256 512; do \
		sips -z $$s $$s $(ICON_SRC) --out "$$iconset/icon_$${s}x$${s}.png" >/dev/null; \
		d=$$((s * 2)); \
		sips -z $$d $$d $(ICON_SRC) --out "$$iconset/icon_$${s}x$${s}@2x.png" >/dev/null; \
	done; \
	iconutil -c icns "$$iconset" -o $(APPDIR)/Contents/Resources/$(BINARY).icns; \
	rm -rf "$$(dirname "$$iconset")"
	@cp $(BINARY) $(APPDIR)/Contents/MacOS/$(BINARY)
	@printf '%s\n' \
		'<?xml version="1.0" encoding="UTF-8"?>' \
		'<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
		'<plist version="1.0">' \
		'<dict>' \
		'  <key>CFBundleName</key><string>Lunarium</string>' \
		'  <key>CFBundleDisplayName</key><string>Lunarium</string>' \
		'  <key>CFBundleIdentifier</key><string>$(BUNDLE_ID)</string>' \
		'  <key>CFBundleExecutable</key><string>$(BINARY)</string>' \
		'  <key>CFBundleIconFile</key><string>$(BINARY)</string>' \
		'  <key>CFBundlePackageType</key><string>APPL</string>' \
		'  <key>CFBundleInfoDictionaryVersion</key><string>6.0</string>' \
		'  <key>CFBundleShortVersionString</key><string>$(APP_VERSION)</string>' \
		'  <key>CFBundleVersion</key><string>$(APP_VERSION)</string>' \
		'  <key>NSHighResolutionCapable</key><true/>' \
		'  <key>LSApplicationCategoryType</key><string>public.app-category.education</string>' \
		'</dict>' \
		'</plist>' > $(APPDIR)/Contents/Info.plist
	@echo ">> built $(APPDIR)"

# Build the bundle and drop it into /Applications.
install-app: app
	@rm -rf "/Applications/$(APP)" && cp -R $(APPDIR) /Applications/ && \
	echo ">> installed /Applications/$(APP)"

clean:
	rm -rf bin/ $(BINARY) $(BINARY).exe

tidy:
	go mod tidy

# Install the OpenGL/X11/ALSA dev headers Ebiten needs to build (cgo) on Linux.
# Not needed on macOS/Windows. Handles Debian (the official golang image) and Fedora.
deps:
	@sudo=""; \
	if [ "$$(id -u)" != 0 ]; then \
		if command -v sudo >/dev/null 2>&1; then sudo="sudo"; \
		else echo "Need root to install packages and 'sudo' is not available — run as root or install the deps manually (see README)."; exit 1; fi; \
	fi; \
	if command -v apt-get >/dev/null 2>&1; then \
		$$sudo apt-get update && $$sudo apt-get install -y --no-install-recommends \
			gcc pkg-config libgl1-mesa-dev libxcursor-dev libxi-dev \
			libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev; \
	elif command -v dnf >/dev/null 2>&1; then \
		$$sudo dnf install -y gcc pkgconf-pkg-config mesa-libGL-devel \
			libXcursor-devel libXi-devel libXinerama-devel libXrandr-devel \
			libXxf86vm-devel alsa-lib-devel; \
	else \
		echo "Unknown package manager — install OpenGL/X11 dev headers manually (see README)"; exit 1; \
	fi

release:
	@bash scripts/release.sh $(filter-out release,$(MAKECMDGOALS))

# The bump level is passed as a goal (`make release patch`), so make also sees
# it as a target — absorb the bump words as no-ops. Listed explicitly (not a
# catch-all `%:`) so genuine typos like `make biuld` still error.
major minor patch breaking feature fix:
	@:
