#import <Cocoa/Cocoa.h>
#include <stdlib.h>
#include <string.h>

// Preferred clipboard image representations, in priority order. Choosing the
// encoded representation when available preserves the original source format.
static NSArray* preferredTypes(void) {
    return @[@"public.png", @"public.jpeg", @"com.compuserve.gif", @"public.tiff"];
}

// Extension codes returned to Go: 1=png 2=jpg 3=gif 4=tiff 0=none.
static int codeForIndex(NSUInteger i) {
    int codes[] = {1, 2, 3, 4};
    return codes[i];
}

// clipboard_image_ext reports the best available image extension without
// copying the bytes. Returns 0 when the clipboard holds no image.
int clipboard_image_ext(void) {
    @autoreleasepool {
        NSPasteboard* pb = [NSPasteboard generalPasteboard];
        NSArray* prefs = preferredTypes();
        for (NSUInteger i = 0; i < prefs.count; i++) {
            if ([pb availableTypeFromArray:@[prefs[i]]] != nil) {
                NSData* d = [pb dataForType:prefs[i]];
                if (d != nil && d.length > 0) {
                    return codeForIndex(i);
                }
            }
        }
        return 0;
    }
}

// clipboard_image_copy returns a malloc'd buffer with the best available image
// representation. The caller must free() it. Sets *outLen and *outExt; returns
// NULL when the clipboard holds no image.
void* clipboard_image_copy(int* outLen, int* outExt) {
    @autoreleasepool {
        NSPasteboard* pb = [NSPasteboard generalPasteboard];
        NSArray* prefs = preferredTypes();
        for (NSUInteger i = 0; i < prefs.count; i++) {
            NSData* d = [pb dataForType:prefs[i]];
            if (d != nil && d.length > 0) {
                *outLen = (int)d.length;
                *outExt = codeForIndex(i);
                void* buf = malloc(d.length);
                if (buf == NULL) {
                    *outLen = 0;
                    *outExt = 0;
                    return NULL;
                }
                memcpy(buf, d.bytes, d.length);
                return buf;
            }
        }
        *outLen = 0;
        *outExt = 0;
        return NULL;
    }
}
