package main

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

void* clipboard_image_copy(int* outLen, int* outExt);
int clipboard_image_ext(void);
*/
import "C"

import "unsafe"

func extName(code C.int) string {
	switch code {
	case 1:
		return "png"
	case 2:
		return "jpg"
	case 3:
		return "gif"
	case 4:
		return "tiff"
	}
	return ""
}

// clipboardImageExt reports the best available image extension, or false when
// the clipboard holds no image.
func clipboardImageExt() (string, bool) {
	code := C.clipboard_image_ext()
	if code == 0 {
		return "", false
	}
	return extName(code), true
}

// clipboardImageData returns the raw bytes of the best available image
// representation and its extension, or false when the clipboard holds no image.
func clipboardImageData() ([]byte, string, bool) {
	var n C.int
	var ext C.int
	ptr := C.clipboard_image_copy(&n, &ext)
	if ptr == nil || n <= 0 {
		return nil, "", false
	}
	defer C.free(unsafe.Pointer(ptr))
	data := C.GoBytes(ptr, n)
	return data, extName(ext), true
}
