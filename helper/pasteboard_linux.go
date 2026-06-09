//go:build linux

package main

import (
	"os"
	"os/exec"
	"strings"
)

// Preferred clipboard MIME types -> file extension, in priority order.
var linuxTypes = []struct {
	mime string
	ext  string
}{
	{"image/png", "png"},
	{"image/jpeg", "jpg"},
	{"image/gif", "gif"},
	{"image/bmp", "bmp"},
	{"image/tiff", "tiff"},
}

// clipboardTool abstracts a CLI clipboard backend (wl-paste or xclip).
type clipboardTool struct {
	listTypes func() []string
	read      func(mime string) ([]byte, bool)
}

func toolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runCmd(name string, args ...string) ([]byte, error) {
	out, err := exec.Command(name, args...).Output()
	return out, err
}

func splitLines(b []byte) []string {
	var out []string
	for _, p := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func wlPasteTool() clipboardTool {
	return clipboardTool{
		listTypes: func() []string {
			out, err := runCmd("wl-paste", "--list-types")
			if err != nil {
				return nil
			}
			return splitLines(out)
		},
		read: func(mime string) ([]byte, bool) {
			out, err := runCmd("wl-paste", "--no-newline", "--type", mime)
			if err != nil || len(out) == 0 {
				return nil, false
			}
			return out, true
		},
	}
}

func xclipTool() clipboardTool {
	return clipboardTool{
		listTypes: func() []string {
			out, err := runCmd("xclip", "-selection", "clipboard", "-t", "TARGETS", "-o")
			if err != nil {
				return nil
			}
			return splitLines(out)
		},
		read: func(mime string) ([]byte, bool) {
			out, err := runCmd("xclip", "-selection", "clipboard", "-t", mime, "-o")
			if err != nil || len(out) == 0 {
				return nil, false
			}
			return out, true
		},
	}
}

// pickTool selects a backend: prefer wl-paste under Wayland, otherwise xclip,
// otherwise whichever is installed.
func pickTool() (clipboardTool, bool) {
	wayland := os.Getenv("WAYLAND_DISPLAY") != ""
	if wayland && toolAvailable("wl-paste") {
		return wlPasteTool(), true
	}
	if toolAvailable("xclip") {
		return xclipTool(), true
	}
	if toolAvailable("wl-paste") {
		return wlPasteTool(), true
	}
	return clipboardTool{}, false
}

func readClipboard() ([]byte, string, bool) {
	tool, ok := pickTool()
	if !ok {
		return nil, "", false
	}

	have := map[string]bool{}
	for _, t := range tool.listTypes() {
		have[t] = true
	}

	// If the backend could enumerate targets, only read advertised types.
	if len(have) > 0 {
		for _, m := range linuxTypes {
			if have[m.mime] {
				if b, ok := tool.read(m.mime); ok {
					return b, m.ext, true
				}
			}
		}
		return nil, "", false
	}

	// Backend could not enumerate; try preferred types directly.
	for _, m := range linuxTypes {
		if b, ok := tool.read(m.mime); ok {
			return b, m.ext, true
		}
	}
	return nil, "", false
}

func clipboardImageExt() (string, bool) {
	_, ext, ok := readClipboard()
	return ext, ok
}

func clipboardImageData() ([]byte, string, bool) {
	return readClipboard()
}
