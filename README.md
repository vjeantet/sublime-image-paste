# Markdown Image Paste (Sublime Text 4 - macOS)

Paste a clipboard image straight into a markdown file.

When the clipboard holds an image and you press `Cmd+V` in a markdown file, the
image is saved next to the current `.md` (or in a configurable subdirectory) as
`<filename>_<N>.<ext>`, and a `![](relative/path)` reference is inserted at the
cursor.

If there is no image in the clipboard (or you are not in a markdown file),
`Cmd+V` keeps its normal paste behavior.

## How it works

- The source format is preserved: `png`, `jpg`, `gif` or `tiff`, depending on
  what the clipboard holds.
- `N` is `(highest existing number) + 1`.
- If the markdown file has not been saved to disk yet, the action is refused
  with a status message (the plugin needs a reference path).

Clipboard access (not native in Sublime Text) is delegated to a small bundled Go
binary, `bin/darwin/imgpaste`, which reads the macOS `NSPasteboard`.

## Installation

### Via Package Control

Install **MarkdownImagePaste** from `Package Control: Install Package`.
(macOS / Sublime Text 4 only.)

### Manual

1. Copy (or symlink) this folder into the Sublime Text `Packages` directory:

   ```sh
   ln -s "$(pwd)" \
     "$HOME/Library/Application Support/Sublime Text/Packages/MarkdownImagePaste"
   ```

2. The `bin/darwin/imgpaste` binary is already precompiled as a **universal**
   binary (arm64 + x86_64); it runs on both Apple Silicon and Intel Macs. To
   rebuild it:

   ```sh
   cd helper
   CGO_ENABLED=1 GOARCH=arm64 go build -o /tmp/imgpaste_arm64 .
   CGO_ENABLED=1 GOARCH=amd64 go build -o /tmp/imgpaste_amd64 .
   lipo -create -output ../bin/darwin/imgpaste /tmp/imgpaste_arm64 /tmp/imgpaste_amd64
   ```

   Or, for the current architecture only:

   ```sh
   cd helper && go build -o ../bin/darwin/imgpaste .
   ```

## Configuration

`Preferences > Package Settings > Markdown Image Paste > Settings`, or edit
`MarkdownImagePaste.sublime-settings`:

```json
{
    // "" = same folder as the .md. Example: "assets".
    "image_subdir": ""
}
```

## Usage

- `Cmd+V` in a markdown file with an image in the clipboard.
- Or via the Command Palette (`Cmd+Shift+P`):
  `PasteImage: Markdown Paste Image From Clipboard`.

## Testing the binary alone

```sh
# with an image in the clipboard
./bin/darwin/imgpaste detect          # -> png|jpg|gif|tiff, exit 0
./bin/darwin/imgpaste save /tmp/x.png # writes the file, exit 0
# without an image (text copied)
./bin/darwin/imgpaste detect          # exit 1
```

## Limitations (v1)

- macOS only (universal arm64 + x86_64 binary). Windows/Linux: keymaps and
  binaries to be added later.
- No image resizing or compression.
