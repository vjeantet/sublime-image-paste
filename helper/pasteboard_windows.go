//go:build windows

package main

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procOpenClipboard           = user32.NewProc("OpenClipboard")
	procCloseClipboard          = user32.NewProc("CloseClipboard")
	procIsClipboardFormatAvail  = user32.NewProc("IsClipboardFormatAvailable")
	procGetClipboardData        = user32.NewProc("GetClipboardData")
	procRegisterClipboardFormat = user32.NewProc("RegisterClipboardFormatW")

	procGlobalLock   = kernel32.NewProc("GlobalLock")
	procGlobalUnlock = kernel32.NewProc("GlobalUnlock")
	procGlobalSize   = kernel32.NewProc("GlobalSize")
)

const (
	cfDIB   = 8  // CF_DIB
	cfDIBV5 = 17 // CF_DIBV5
)

func registerFormat(name string) uint32 {
	p, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0
	}
	r, _, _ := procRegisterClipboardFormat.Call(uintptr(unsafe.Pointer(p)))
	return uint32(r)
}

func isAvailable(format uint32) bool {
	r, _, _ := procIsClipboardFormatAvail.Call(uintptr(format))
	return r != 0
}

// clipboardBytes copies the bytes behind a clipboard format handle. The handle
// stays owned by the clipboard (not freed here).
func clipboardBytes(format uint32) []byte {
	h, _, _ := procGetClipboardData.Call(uintptr(format))
	if h == 0 {
		return nil
	}
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return nil
	}
	defer procGlobalUnlock.Call(h)
	size, _, _ := procGlobalSize.Call(h)
	if size == 0 {
		return nil
	}
	buf := make([]byte, int(size))
	// `ptr` points at OS-owned global memory locked by GlobalLock; it is not
	// Go-managed, so the uintptr->Pointer conversion is stable here (go vet's
	// "possible misuse" heuristic does not apply to non-Go memory).
	copy(buf, unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(size)))
	return buf
}

// readClipboard returns the best available image bytes and its extension.
// Encoded formats (PNG/JPEG/GIF) registered by the source app are returned
// as-is to preserve the source format; a device-independent bitmap is wrapped
// into a .bmp file.
func readClipboard() ([]byte, string, bool) {
	if r, _, _ := procOpenClipboard.Call(0); r == 0 {
		return nil, "", false
	}
	defer procCloseClipboard.Call()

	encoded := []struct {
		name string
		ext  string
	}{
		{"PNG", "png"},
		{"image/png", "png"},
		{"JFIF", "jpg"},
		{"image/jpeg", "jpg"},
		{"GIF", "gif"},
		{"image/gif", "gif"},
	}
	for _, f := range encoded {
		id := registerFormat(f.name)
		if id != 0 && isAvailable(id) {
			if b := clipboardBytes(id); len(b) > 0 {
				return b, f.ext, true
			}
		}
	}

	for _, cf := range []uint32{cfDIBV5, cfDIB} {
		if isAvailable(cf) {
			if dib := clipboardBytes(cf); len(dib) >= 40 {
				return dibToBMP(dib), "bmp", true
			}
		}
	}
	return nil, "", false
}

// dibToBMP prepends a 14-byte BITMAPFILEHEADER to a DIB so the result is a
// valid standalone .bmp file.
func dibToBMP(dib []byte) []byte {
	biSize := binary.LittleEndian.Uint32(dib[0:4])
	bitCount := binary.LittleEndian.Uint16(dib[14:16])
	compression := binary.LittleEndian.Uint32(dib[16:20])
	clrUsed := binary.LittleEndian.Uint32(dib[32:36])

	var paletteBytes uint32
	if bitCount <= 8 {
		n := clrUsed
		if n == 0 {
			n = uint32(1) << bitCount
		}
		paletteBytes = n * 4
	}
	var maskBytes uint32
	if compression == 3 && biSize == 40 { // BI_BITFIELDS with a BITMAPINFOHEADER
		maskBytes = 12
	}
	offBits := 14 + biSize + paletteBytes + maskBytes

	header := make([]byte, 14)
	header[0] = 'B'
	header[1] = 'M'
	binary.LittleEndian.PutUint32(header[2:6], uint32(14+len(dib)))
	binary.LittleEndian.PutUint32(header[10:14], offBits)

	out := make([]byte, 0, 14+len(dib))
	out = append(out, header...)
	out = append(out, dib...)
	return out
}

func clipboardImageExt() (string, bool) {
	_, ext, ok := readClipboard()
	return ext, ok
}

func clipboardImageData() ([]byte, string, bool) {
	return readClipboard()
}
