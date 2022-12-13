package main

// #cgo pkg-config: gdk-3.0 atk gtk+-3.0
// #include "tray.go.h"
import "C"

import "github.com/gotk3/gotk3/gtk"

func onStatusIconSetup(window *gtk.ApplicationWindow) {
	C.status_icon_setup((C.gpointer)(window.ToWidget().Native()))
	setStatusIconText(appName)
}

func setStatusIconText(text string) {
	C.status_icon_text_set(C.CString(text))
}

func setStatusIconHide() {
	C.status_icon_hide()
}
