//go:build linux

package main

/*
#cgo LDFLAGS: -lX11 -lXi
#include <X11/Xlib.h>
#include <X11/extensions/XInput2.h>

void BlockInputs() {
    Display *display = XOpenDisplay(NULL);
    Window root = DefaultRootWindow(display);

    // Block mouse
    XGrabPointer(display, root, True, ButtonPressMask | ButtonReleaseMask,
                 GrabModeAsync, GrabModeAsync, None, None, CurrentTime);

    // Block keyboard (XInput2 method)
    XIEventMask evmask;
    unsigned char mask[(XI_LASTEVENT + 7)/8] = { 0 };
    evmask.deviceid = XIAllDevices;
    evmask.mask_len = sizeof(mask);
    evmask.mask = mask;
    XISetMask(mask, XI_KeyPress);
    XISetMask(mask, XI_KeyRelease);

    XISelectEvents(display, root, &evmask, 1);
}

void UnblockInputs() {
    Display *display = XOpenDisplay(NULL);
    XUngrabPointer(display, CurrentTime);
}
*/
import "C"

func BlockInputs() {
	C.BlockInputs()
}

func UnblockInputs() {
	C.UnblockInputs()
}
