package main

import (
	"errors"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
// #include "icon.go.h"
import "C"

func getIcon() (*gdk.Pixbuf, error) {

	c := C.gdk_pixbuf_new_from_xpm_data(&C.thlink_client_gtk_xpm[0])
	if c == nil {
		return nil, errors.New("get icon error")
	}
	obj := &glib.Object{GObject: glib.ToGObject(unsafe.Pointer(c))}
	p := &gdk.Pixbuf{Object: obj}

	return p, nil
}
