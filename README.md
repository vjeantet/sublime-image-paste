# Markdown Image Paste (Sublime Text 4 - macOS & Windows)

Paste a clipboard image straight into a markdown file.

When the clipboard holds an image and you press the paste shortcut (`Cmd+V` on
macOS, `Ctrl+V` on Windows) in a markdown file, the image is saved next to the
current `.md` (or in a configurable subdirectory) as `<filename>_<N>.<ext>`, and
a `![](relative/path)` reference is inserted at the cursor.

If there is no image in the clipboard (or you are not in a markdown file), the
paste shortcut keeps its normal behavior.

## How it works

- The source format is preserved when possible: `png`, `jpg`, `gif` or `tiff`.
  On Windows, when the clipboard only holds a raw bitmap (e.g. a screenshot),
  the image is saved as `bmp`.
- `N` is `(highest existing number) + 1`.
- If the markdown file has not been saved to disk yet, the action is refused
  with a status message (the plugin needs a reference path).

Clipboard access (not native in Sublime Text) is delegated to a small bundled Go
binary, selected per platform:

- macOS: `bin/darwin/imgpaste`, reads `NSPasteboard`.
- Windows: `bin/windows/imgpaste.exe`, reads the Win32 clipboard.

## Installation

### Via Package Control

Install **MarkdownImagePaste** from `Package Control: Install Package`.
(Sublime Text 4, macOS or Windows.)

### Manual

Copy (or symlink) this folder into the Sublime Text `Packages` directory, e.g.
on macOS:

```sh
ln -s "$(pwd)" \
  "$HOME/Library/Application Support/Sublime Text/Packages/MarkdownImagePaste"
```

The prebuilt helper binaries are already bundled:

- `bin/darwin/imgpaste` - universal macOS binary (arm64 + x86_64).
- `bin/windows/imgpaste.exe` - Windows x86-64 binary.

To rebuild them:

```sh
cd helper

# macOS universal binary (run on macOS)
CGO_ENABLED=1 GOARCH=arm64 go build -o /tmp/imgpaste_arm64 .
CGO_ENABLED=1 GOARCH=amd64 go build -o /tmp/imgpaste_amd64 .
lipo -create -output ../bin/darwin/imgpaste /tmp/imgpaste_arm64 /tmp/imgpaste_amd64

# Windows binary (cross-compiles from any OS, no cgo)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ../bin/windows/imgpaste.exe .
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

- Paste shortcut (`Cmd+V` on macOS, `Ctrl+V` on Windows) in a markdown file with
  an image in the clipboard.
- Or via the Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`):
  `PasteImage: Markdown Paste Image From Clipboard`.

## Testing the binary alone

```sh
# with an image in the clipboard (macOS shown; on Windows use imgpaste.exe)
./bin/darwin/imgpaste detect          # -> png|jpg|gif|tiff|bmp, exit 0
./bin/darwin/imgpaste save /tmp/x.png # writes the file, exit 0
# without an image (text copied)
./bin/darwin/imgpaste detect          # exit 1
```

## Limitations

- macOS (universal arm64 + x86_64) and Windows (x86-64). Linux: keymap and
  binary to be added later.
- On Windows, images available only as a raw bitmap are saved as `.bmp`
  (lossless, but larger and not previewed by every markdown renderer).
- No image resizing or compression.
