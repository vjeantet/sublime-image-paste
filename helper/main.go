// Command imgpaste reads an image from the macOS clipboard (NSPasteboard) and
// either reports its format (detect) or writes its raw bytes to a file (save).
//
// Usage:
//
//	imgpaste detect            # prints extension (png/jpg/gif/tiff), exit 0 if image present, 1 otherwise
//	imgpaste save <outfile>    # writes the best image representation to <outfile>, prints extension, exit 0/1
//
// The original encoded bytes are preserved when the clipboard exposes an
// encoded representation (PNG/JPEG/GIF). For TIFF-only clipboards (common when
// copying from screenshots or some apps), the raw TIFF bytes are written.
package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: imgpaste detect | imgpaste save <outfile>")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "detect":
		ext, ok := clipboardImageExt()
		if !ok {
			os.Exit(1)
		}
		fmt.Println(ext)
		os.Exit(0)

	case "save":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		out := os.Args[2]
		data, ext, ok := clipboardImageData()
		if !ok {
			fmt.Fprintln(os.Stderr, "no image in clipboard")
			os.Exit(1)
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "write error:", err)
			os.Exit(1)
		}
		fmt.Println(ext)
		os.Exit(0)

	default:
		usage()
		os.Exit(2)
	}
}
