package glgo

import (
	"errors"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include "glg_cairo.h"

GlgLineGraph * my_glg_line_graph_new()
{
	return glg_line_graph_new ("chart-set-elements",
			GLG_TOOLTIP | GLG_TITLE_T | GLG_TITLE_X | GLG_TITLE_Y | GLG_GRID_MAJOR_X |  GLG_GRID_MAJOR_Y | GLG_GRID_MINOR_X |  GLG_GRID_MINOR_Y | GLG_GRID_LABELS_X | GLG_GRID_LABELS_Y,
			"range-tick-minor-x", 1,
			"range-tick-major-x", 10,
			"range-scale-minor-x", 0,
			"range-scale-major-x", 40,
			"range-tick-minor-y", 2,
			"range-tick-major-y", 10,
			"range-scale-minor-y", 0,
			"range-scale-major-y", 120,
			"series-line-width", 2,
			"graph-title-foreground",  "black",
			"graph-scale-foreground",  "black",
			"graph-chart-background",  "light gray",
			"graph-window-background", "white",
			"text-title-main", "<big><b>Tunnel Delay Line Chart</b></big>",
			"text-title-yaxis", "<span>delay(ms)</span>",
			"text-title-xaxis", "<i>Click mouse button 1 to <span foreground=\"red\">toggle</span> popup legend.</i>",
			NULL);
}
*/
import "C"

type GlgLineGraph struct {
	gtk.Bin

	// rangeScaleMajorY int
}

func GlgLineGraphNew() (*GlgLineGraph, error) {
	glg := C.my_glg_line_graph_new()
	if glg == nil {
		return nil, errors.New("cgo returned unexpected nil pointer")
	}

	obj := glib.Take(unsafe.Pointer(glg))

	return &GlgLineGraph{Bin: gtk.Bin{Container: gtk.Container{Widget: gtk.Widget{InitiallyUnowned: glib.InitiallyUnowned{Object: obj}}}}}, nil
}

func (g *GlgLineGraph) GlgLineGraphDataSeriesAdd(legend string, color string) bool {

	cLegend := C.CString(legend)
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cLegend))
	defer C.free(unsafe.Pointer(cColor))

	return C.glg_line_graph_data_series_add((*C.GlgLineGraph)(unsafe.Pointer(g.Native())), cLegend, cColor) == C.TRUE
}

func (g *GlgLineGraph) GlgLineGraphDataSeriesAddValue(series int, value float64) bool {
	defer g.glgLineGraphRedraw()

	// auto scale
	/*if value > float64(g.rangeScaleMajorY) {
		g.glgLineGraphChartSetYRanges((int(math.Floor(value/10)) + 1) * 10)
	}*/

	return C.glg_line_graph_data_series_add_value((*C.GlgLineGraph)(unsafe.Pointer(g.Native())), *(*C.int)(unsafe.Pointer(&series)), *(*C.double)(unsafe.Pointer(&value))) == C.TRUE
}

func (g *GlgLineGraph) glgLineGraphRedraw() {
	C.glg_line_graph_redraw((*C.GlgLineGraph)(unsafe.Pointer(g.Native())))
}

// cannot set range more than once
/*func (g *GlgLineGraph) glgLineGraphChartSetYRanges(yScaleMax int) {
	C.glg_line_graph_chart_set_y_ranges((*C.GlgLineGraph)(unsafe.Pointer(g.Native())), 2, 10, 0, *(*C.gint)(unsafe.Pointer(&yScaleMax)))
}*/
