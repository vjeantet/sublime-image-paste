# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Sublime Text 4 package (`MarkdownImagePaste`) that pastes a clipboard image into a markdown file: it saves the image next to the `.md` as `<mdstem>_<NN>.<ext>` and inserts a `![](relative/path)` reference. Two parts:

- A Python plugin (`MarkdownImagePaste.py`) running in Sublime's plugin host (Python 3.8).
- A standalone Go helper (`helper/`) compiled per platform into `bin/{darwin,linux,windows}/`, because Sublime has no native clipboard-image access.

## Architecture

The plugin and helper communicate over a tiny CLI contract, not a library boundary:

- `imgpaste detect` -> prints the extension (`png`/`jpg`/`gif`/`tiff`/`bmp`) and exits 0 if the clipboard holds an image, exits 1 otherwise.
- `imgpaste save <outfile>` -> writes the raw image bytes to `<outfile>`, prints the extension, exits 0/1.

`MdPasteImageCommand.run` (bound to the platform paste key only inside `text.html.markdown` scope) orchestrates: refuse if the view is unsaved -> `detect` -> on no-image fall back to the built-in `paste` command -> compute the next numeric suffix by scanning the target dir (`_next_index`) -> `save` -> insert the relative reference. The key design choice is that a missing image or any helper error degrades gracefully to a normal paste, so the shortcut never feels broken.

The Go helper is split by build tag, each file implementing the same three functions (`clipboardImageExt`, `clipboardImageData`, and `readClipboard`):

- `pasteboard_darwin.go` + `pasteboard_darwin.m` - cgo over `NSPasteboard` (requires `CGO_ENABLED=1`, `-framework Cocoa`).
- `pasteboard_windows.go` - Win32 clipboard via `syscall` (no cgo); wraps a raw DIB into a `.bmp` when only a bitmap is present.
- `pasteboard_linux.go` - shells out to `wl-paste` (Wayland) or `xclip` (X11); no cgo. Requires one of those tools installed at runtime.

Encoded clipboard formats (PNG/JPEG/GIF) are written through unchanged to preserve the source format; only raw-bitmap-only clipboards get re-wrapped (BMP on Windows). Keep this format-preservation behavior when touching the helpers.

## Building the helpers

Always rebuild from `helper/` and output into the committed `bin/` paths (the prebuilt binaries are checked in and shipped with the package):

```sh
cd helper

# macOS universal binary (must run on macOS for cgo)
CGO_ENABLED=1 GOARCH=arm64 go build -o /tmp/imgpaste_arm64 .
CGO_ENABLED=1 GOARCH=amd64 go build -o /tmp/imgpaste_amd64 .
lipo -create -output ../bin/darwin/imgpaste /tmp/imgpaste_arm64 /tmp/imgpaste_amd64

# Windows (cross-compiles anywhere, no cgo)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ../bin/windows/imgpaste.exe .

# Linux (cross-compiles anywhere, no cgo; GOARCH=arm64 for ARM)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../bin/linux/imgpaste .
```

The macOS and Linux binaries are shipped without an exec bit; the plugin chmods them at runtime (`_ensure_executable`).

## Testing the helper directly

```sh
# with an image copied to the clipboard
./bin/darwin/imgpaste detect            # -> png|jpg|gif|tiff|bmp, exit 0
./bin/darwin/imgpaste save /tmp/x.png   # writes the file, exit 0
# with text (no image) copied
./bin/darwin/imgpaste detect            # exit 1
```

There is no automated test suite. The plugin itself can only be exercised inside Sublime Text 4: symlink the repo into the `Packages` dir, reload, and paste an image into a saved `.md`.

## Conventions

- Keymaps are per-platform (`Default (OSX|Windows|Linux).sublime-keymap`), each binding the native paste key to `md_paste_image` scoped to markdown. The macOS uses `super+v`. Edit all three together when changing the binding.
- Project resources (README, comments, commit messages) are written in English. Note: the Python status messages currently contain French strings - prefer English for any new user-facing text.
- New work goes on a feature branch; do not push, tag, or release without an explicit request.
