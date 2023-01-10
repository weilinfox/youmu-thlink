/* glg_cairo.c
 * ----------------------------------------------
 *
 *
 * A GTK+ widget that implements a modified XY line graph.
 * - Y points are plotted
 * - X point on all series is implied - by entry order
 * - X-scale rolls to display the most recent data (i.e. show last 40 points)
 * - Support unlimited number of data series
 * - Popup legend via mouse-button one
 * - Supports X Y & Page Titles
 *   
 *
 * (c) 2007, 2016 James Scott Jr
 *
 * Authors:
 *   James Scott Jr <skoona@gmail.com>
 *
 * (c) 2023, 2023 weilinfox
 * Date: 1/2023
 *
 * Contributors:
 *   weilinfox <weilinfox@inuyasha.love>
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this library; if not, see <https://www.gnu.org/licenses/>.
 */

 
/**
 * SECTION:glg_cairo
 * <title>A GTK/cairo Line Graph Widget.</title> 
 * @short_description: A simple xy line graph widget written using only GTK and cairo.
 * @see_also: #GlgLineGraph
 * @stability: Stable
 * @include: glg_cairo.h
 *
 * <mediaobject>
 *   <imageobject>
 *     <imagedata fileref="glg_cairo3.png" format="PNG"/>
 *   </imageobject>
 * </mediaobject>
 *
 *    A Gtk+3/GLib2/Cairo GUI application which demonstrates the use of GTK+ and cairo for
 * producing xy line graphs.  This widget once created allows you to add one or more
 * data series, then add values to those series for plotting.  The X point is assumed based
 * on arrival order.  However, the Y value or position is based on the current scale and
 * the y value itself.  If the charts x scale maximum is 40, or 40 points, the 41+ value is
 * appended to the 40th position after pos 0 is dropped - effectively rolling the x points
 * from right to left in the chart view.
 *
 * <emphasis>FEATURES</emphasis>
 * <itemizedlist>
 *  <listitem>
 *  Unlimited data series support.
 *  </listitem>
 *  <listitem>
 *  Accurate scaling across a wide range of X & Y scales.
 *  </listitem>
 *  <listitem>
 *  Using values ranges above or below 1.
 *  </listitem>
 *  <listitem>
 *  Rolling data points, if number of x points exceed x-scale. (left shift)
 *  </listitem>
 *  <listitem>
 *  Ability to change chart background color, window backgrounds colors, etc.
 *  </listitem>
 *  <listitem>
 *  Popup Tooltip, via mouse-button-1 click to enable/toggle. Tooltip overlays top graph title, when present.
 *  </listitem>
 *  <listitem>
 *  Data points are time stamped with current time when added.
 *  </listitem>
 *  <listitem>
 *  Auto Size to current window size; i.e. no-scrolling.
 *  </listitem>
 * </itemizedlist>
 *
 * Packaged as a gtk widget for ease of use and easy inclusion into new or existing programs.
 * 
 * The GlgLinegraph widget has a gobject property for every control option except creating a 
 * new data series glg_line_graph_series_add(), and adding a value to that series with
 * glg_line_graph_series_add_value().
 * 
 * One signal is available 'point-selected' which outputs the Y value most likely under the 
 * mouse ptr.  For correlation purposes the position of the mouse and the position of the Y point
 * is given in case two or more points are returned.
 *
 * The scale or range of the chart is dependant on the Y-values.  For value greater than 1, the range/scale
 * should be set to whole numbers say 0 to 100.  For values less than 1, use a range/scale of 0 to 1 .
 *  
 * The following api's will create a version of this line graph:
 * <example>
 *  <title>Using a GlgLineGraph with gobject methods.</title>
 *  <programlisting>
 *  #include <gtk/gtk.h>
 *  #include <glg_cairo.h>
 *  ...
 *  GlgLineGraph *glg = NULL;
 *  gint  i_series_0 = 0, i_series_1 = 0;
 *  ...
 *  glg = glg_line_graph_new(
 *              "range-tick-minor-x", 1,
 *   					"range-tick-major-x", 2,
 *   					"range-scale-minor-x", 0,
 *   					"range-scale-major-x", 40,
 *   					"range-tick-minor-y", 5,
 *   					"range-tick-major-y", 10,
 *   					"range-scale-minor-y", 0,
 *   					"range-scale-major-y", 100,
 *   					"chart-set-elements", GLG_TOOLTIP | 
 *						GLG_GRID_LABELS_X | GLG_GRID_LABELS_Y |
 *                	 	GLG_TITLE_T | GLG_TITLE_X | GLG_TITLE_Y |                                     
 *                	 	GLG_GRID_LINES | GLG_GRID_MINOR_X | GLG_GRID_MAJOR_X |
 *                    	GLG_GRID_MINOR_Y | GLG_GRID_MAJOR_Y, 
 *                    	"series-line-width", 3, 
 *                      "graph-title-foreground",  "blue",
 *   					"graph-scale-foreground",  "red",
 *    					"graph-chart-background",  "light blue",
 *    					"graph-window-background", "white", 
 *    					"text-title-main", "This Top Title Line ",        				
 *                      "text-title-yaxis", "This is the Y axis title line.",
 *         				"text-title-xaxis", "This is the X axis title line.",
 *   			        NULL);
 *
 *  gtk_container_add (GTK_CONTAINER (window), GTK_WIDGET(glg));
 *  gtk_widget_show_all (window);
 *
 *  i_series_0 = glg_line_graph_data_series_add (glg, "Volts", "red");
 *  i_series_1 = glg_line_graph_data_series_add (glg, "Battery", "blue");
 *
 *  glg_line_graph_data_series_add_value (glg, i_series_0, 66.0);
 *  glg_line_graph_data_series_add_value (glg, i_series_0, 73.0);
 *
 *  glg_line_graph_data_series_add_value (glg, i_series_1, 56.8);
 *  glg_line_graph_data_series_add_value (glg, i_series_1, 83.6);
 *  
 *
 *  glg_line_graph_redraw ( graph );
 *   	   
 *  </programlisting>
 * </example>
 * 
 * Or the following standard api method will also work.
 * <example>
 *  <title>Using a GlgLineGraph with standard APIs.</title>
 *  <programlisting>
 *  #include <gtk/gtk.h>
 *  #include <glg_cairo.h>
 *  ...
 * GlgLineGraph *glg = NULL;
 * gint  i_series_0 = 0, i_series_1 = 0;
 * ...
 * glg = glg_line_graph_new(NULL);
 *
 * glg_line_graph_chart_set_x_ranges (glg, 1, 2,0, 40);
 * glg_line_graph_chart_set_y_ranges (glg, 5,10,0,100);
 *
 * glg_line_graph_chart_set_elements (glg, GLG_TOOLTIP | 
 *				        GLG_GRID_LABELS_X | GLG_GRID_LABELS_Y |
 *               	 	GLG_TITLE_T | GLG_TITLE_X | GLG_TITLE_Y |                                     
 *               	 	GLG_GRID_LINES | GLG_GRID_MINOR_X | GLG_GRID_MAJOR_X |
 *                   	GLG_GRID_MINOR_Y | GLG_GRID_MAJOR_Y 
 *                   			     ); 
 *
 * glg_line_graph_chart_set_text (glg, GLG_TITLE_T, "This Top Title Line " );	
 *        				
 * glg_line_graph_chart_set_text (glg, GLG_TITLE_Y, "This is the y label." ); 	
 *        					 
 * glg_line_graph_chart_set_text (glg, GLG_TITLE_X, "This is the x label" );
 *
 * glg_line_graph_chart_set_color (glg, GLG_TITLE,  "blue");
 * glg_line_graph_chart_set_color (glg, GLG_SCALE,  "read");
 * glg_line_graph_chart_set_color (glg, GLG_CHART,  "light blue");
 * glg_line_graph_chart_set_color (glg, GLG_WINDOW, "white");
 * 
 * gtk_container_add (GTK_CONTAINER (window), GTK_WIDGET(glg));
 * gtk_widget_show_all (window);
 *
 * i_series_0 = glg_line_graph_data_series_add (glg, "Volts", "red");
 * i_series_1 = glg_line_graph_data_series_add (glg, "Battery", "blue");
 *
 * glg_line_graph_data_series_add_value (glg, i_series_0, 66.0);
 * glg_line_graph_data_series_add_value (glg, i_series_0, 73.0);
 *
 * glg_line_graph_data_series_add_value (glg, i_series_1, 56.8);
 * glg_line_graph_data_series_add_value (glg, i_series_1, 83.6);
 *
 * glg_line_graph_redraw ( glg );
 * 	   
 *  </programlisting>
 * </example>
 *     
 *
 * A 'lgcairo.c' demonstration proram is included to illustrate how to quickly use the widget.
 *
 * <note>
 *  <title>lgcairo.c demonstration program</title>
 *  <link linkend="lgcairo-Sample-Program">C example program</link>
 * </note>
 * 
 */

#include <math.h>
#include <time.h>
#include <gtk/gtk.h>

#include "glg_cairo.h"

#define GLG_MAX_BUFFER  512      /* Size of a text buffer or local string */


/*
 * Individual Data Series for Plotting
*/
typedef struct _GLG_SERIES {
    gint        cb_id;
    gint        i_series_id;    /* is this series number 1 2 or 3, ZERO based */
    gint        i_point_count;  /* 1 based */
    gint        i_max_points;   /* 1 based */
    gchar       ch_legend_text[GLG_MAX_STRING];
    gchar       ch_legend_color[GLG_MAX_STRING];
    GdkRGBA    legend_color;
    gdouble     d_max_value;
    gdouble     d_min_value;
    gdouble    *lg_point_dvalue;    /* array of doubles y values zero based, x = index */
    GdkPoint   *point_pos;      /* last gdk position each point - recalc on evey draw */
} GLG_SERIES, *PGLG_SERIES;

/*
 * Chart dimensions for drawing chart box
*/
typedef struct _GLG_RANGES {
    gint        cb_id;
    gint        i_inc_minor_scale_by;   /* minor increments */
    gint        i_inc_major_scale_by;   /* major increments */
    gint        i_min_scale;    /* minimum scale value - ex:   0 */
    gint        i_max_scale;    /* maximum scale value - ex: 100 */
    gint        i_num_minor;    /* number of minor points */
    gint        i_num_major;    /* number of major points */
    gint        i_minor_inc;    /* pixels per minor increment */
    gint        i_major_inc;    /* pixels per major increment */
} GLG_RANGE , *PGLG_RANGE;

/* 
 * widget private data structure 
*/
struct _GlgLineGraphPrivate
{
    gint        cb_id;			  /* structure id */	
	GdkWindow   *window;
    GLGElementID lgflags;         /* things to be drawn */
    /* new cairo design */
    cairo_surface_t *surface;
    cairo_t 	 *cr;    
    cairo_rectangle_int_t page_title_box;
    cairo_rectangle_int_t tooltip_box;
    cairo_rectangle_int_t x_label_box;
    cairo_rectangle_int_t y_label_box;
    cairo_rectangle_int_t plot_box;      /* actual size of graph area */
    cairo_rectangle_int_t page_box;      /* entire window size */
    /* element colors */
    GdkRGBA    window_color;    /* actual gdk color -- needs to be color/65535 to match cairo scale 0-1.0  */
    GdkRGBA    chart_color;
    GdkRGBA    scale_color;
    GdkRGBA    title_color;
    GdkRGBA    series_color;
    /* mouse device */
    /* GdkDeviceManager *device_manager; */
    GdkDevice        *device_pointer;
    /* data points and tooltip info */
    gint        i_points_available;
    gint        i_num_series;   /* 1 based */
    GList      *lg_series;      /* double-linked list of data series PGLG_SERIES */
    GList      *lg_series_time; /* time_t of each sample */
    gint       series_line_width;        /* drawn line width for data series -- default: 3 */
    /* buffer around all sides */
    gint        x_border;
    gint        y_border;
    gint		xfactor;               /* default pixel size of one 'M' */
    gint		yfactor;			   /* default pixel size of one 'M' */
    /* current mouse position */
    gboolean    b_tooltip_active;
    gboolean    b_mouse_onoff;
    GdkPoint    mouse_pos;
    GdkModifierType mouse_state;
    /* color names, labels, and titles */
    gchar       ch_color_window_bg[GLG_MAX_STRING];
    gchar       ch_color_chart_bg[GLG_MAX_STRING];
    gchar       ch_color_title_fg[GLG_MAX_BUFFER];
    gchar       ch_color_scale_fg[GLG_MAX_STRING];
    gchar       ch_tooltip_text[GLG_MAX_BUFFER];
    gchar      *x_label_text;
    gchar      *y_label_text;
    gchar      *page_title_text;
    /* chart scales */
    GLG_RANGE    x_range;   /* scale mechanics */
    GLG_RANGE    y_range;	/* scale mechanics */
};

/* 
 * Internal working structure ids 
*/
typedef enum _GLG_Control_Block_id {
    GLG_NO_ID,
    GLG_SERIES_ID,
    GLG_RANGE_ID,
    GLG_GRAPH_ID,
    GLG_PRIVATE_ID,    
    GLG_NUM_ID
} GLGDataID;

/* 
 * Internal working property ids 
*/
enum _GLG_PROPERTY_ID {
  PROP_0,
  PROP_GRAPH_DRAWING_TYPE,
  PROP_GRAPH_TITLE,
  PROP_GRAPH_TITLE_X,
  PROP_GRAPH_TITLE_Y,
  PROP_GRAPH_LINE_WIDTH,
  PROP_GRAPH_ELEMENTS,
  PROP_GRAPH_TITLE_COLOR,      
  PROP_GRAPH_SCALE_COLOR,      
  PROP_GRAPH_CHART_COLOR,      
  PROP_GRAPH_WINDOW_COLOR,
  PROP_TICK_MINOR_X,
  PROP_TICK_MAJOR_X,
  PROP_SCALE_MINOR_X,
  PROP_SCALE_MAJOR_X,
  PROP_TICK_MINOR_Y,
  PROP_TICK_MAJOR_Y,
  PROP_SCALE_MINOR_Y,
  PROP_SCALE_MAJOR_Y
} GLG_PROPERTY_ID;

G_DEFINE_TYPE_WITH_PRIVATE (GlgLineGraph, glg_line_graph, GTK_TYPE_WIDGET);

#define GLG_LINE_GRAPH_GET_PRIVATE(obj) (G_TYPE_INSTANCE_GET_PRIVATE ((obj), GLG_TYPE_LINE_GRAPH, GlgLineGraphPrivate))
#define GLG_LINE_GRAPH_CLASS(obj)   (G_TYPE_CHECK_CLASS_CAST ((obj), GLG_LINE_GRAPH, GlgLineGraphClass))
#define GLG_IS_LINE_GRAPH_CLASS(obj)    (G_TYPE_CHECK_CLASS_TYPE ((obj), GLG_TYPE_LINE_GRAPH))
#define GLG_LINE_GRAPH_GET_CLASS    (G_TYPE_INSTANCE_GET_CLASS ((obj), GLG_TYPE_LINE_GRAPH, GlgLineGraphClass))

/*
 * Private routines for graph widget internal functions
*/
gint64 glg_duration_us(gint64 *start_time, gchar *method_name); /* utils */

static void 	glg_line_graph_class_init (GlgLineGraphClass *klass);
static void glg_line_graph_get_property (GObject *object, guint prop_id, GValue *value, GParamSpec *pspec);
static void glg_line_graph_set_property (GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec);
static gint glg_line_graph_data_series_remove_all (GlgLineGraph *graph);
static void 	glg_line_graph_destroy (GtkWidget *object);

static void     glg_line_graph_realize (GtkWidget *widget);
static void     glg_line_graph_size_allocate(GtkWidget *widget, GtkAllocation *allocation);
static void     glg_line_graph_send_configure (GlgLineGraph *graph);
static gboolean glg_line_graph_configure_event (GtkWidget *widget, GdkEventConfigure *event);
static gboolean glg_line_graph_compute_layout(GlgLineGraph *graph, GdkRectangle *allocation);
static gboolean glg_line_graph_master_draw (GtkWidget *graph, cairo_t *cr);

static gboolean glg_line_graph_button_press_event (GtkWidget * widget, GdkEventButton * ev);
static gboolean glg_line_graph_motion_notify_event (GtkWidget * widget, GdkEventMotion * ev);

static void 	glg_line_graph_draw_graph (GtkWidget *graph);
static gint 	glg_line_graph_draw_tooltip (GlgLineGraph *graph);
static void 	glg_line_graph_draw_x_grid_labels (GlgLineGraph *graph);
static void 	glg_line_graph_draw_y_grid_labels (GlgLineGraph *graph);
static gint 	glg_line_graph_draw_text_horizontal (GlgLineGraph *graph, gchar * pch_text, cairo_rectangle_int_t * rect);
static gint 	glg_line_graph_draw_text_vertical (GlgLineGraph *graph, gchar * pch_text, cairo_rectangle_int_t * rect);
static gint 	glg_line_graph_draw_grid_lines (GlgLineGraph *graph);
static gint 	glg_line_graph_data_series_draw (GlgLineGraph *graph, PGLG_SERIES psd);
static gint 	glg_line_graph_data_series_draw_all (GlgLineGraph *graph, gboolean redraw_control);

static void _glg_cairo_marshal_VOID__DOUBLE_DOUBLE_DOUBLE_DOUBLE (GClosure     *closure,
                                                      GValue       *return_value,
                                                      guint         n_param_values,
                                                      const GValue *param_values,
                                                      gpointer      invocation_hint,
                                                      gpointer      marshal_data);

enum
{
	POINT_SELECTED_SIGNAL,
	LAST_SIGNAL
};

static guint glg_line_graph_signals[LAST_SIGNAL] = { 0 };

/*
 * Function definitions
*/
static void glg_line_graph_class_init (GlgLineGraphClass *klass)
{
	GObjectClass 	*obj_class 		 = G_OBJECT_CLASS (klass);
	GtkWidgetClass  *widget_class 	 = GTK_WIDGET_CLASS (klass);

    gint			elements = GLG_GRID_LINES;      

    g_debug ("===> glg_line_graph_class_init(entered)");

    /* GObject signal overrides */
    obj_class->set_property = glg_line_graph_set_property;
    obj_class->get_property = glg_line_graph_get_property;

	/* GtkWidget signals overrides */
    widget_class->realize               = glg_line_graph_realize;
	widget_class->configure_event 	    = glg_line_graph_configure_event;
	widget_class->draw      			= glg_line_graph_master_draw;
	widget_class->motion_notify_event   = glg_line_graph_motion_notify_event;
  	widget_class->button_press_event    = glg_line_graph_button_press_event;
    widget_class->size_allocate         = glg_line_graph_size_allocate;
    widget_class->destroy               = glg_line_graph_destroy;

	/**
	 * GlgLineGraph::point-selected:
	 * @widget: the line graph widget that received the signal
	 * @x_value: x value on the chart
	 * @y_value: y value on the chart
	 * @point_y_pos: pixel position of y value on chart
	 * @mouse_y_pos: pixel position of the mouse ptr on chart
	 *
	 * The ::point-selected signal is emitted after the toggle-on mouse1 click, and sends
	 * values closest to the mouse pointer.  
	 */
     glg_line_graph_signals[POINT_SELECTED_SIGNAL] = g_signal_new (
			"point-selected",
			G_OBJECT_CLASS_TYPE (obj_class),
			G_SIGNAL_RUN_FIRST,
			G_STRUCT_OFFSET (GlgLineGraphClass, point_selected),
			NULL, NULL,
			_glg_cairo_marshal_VOID__DOUBLE_DOUBLE_DOUBLE_DOUBLE,
			G_TYPE_NONE, 4,
			G_TYPE_DOUBLE,
			G_TYPE_DOUBLE,
			G_TYPE_DOUBLE,
			G_TYPE_DOUBLE);


    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_TITLE,
                                   g_param_spec_string ("text-title-main",
                                                        "Graph Top Title",
                                                        "Title at top of graph on the X axis",
                                                        "<big><b>Top Title</b></big>",
                                                        G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_TITLE_X,
                                   g_param_spec_string ("text-title-xaxis",
                                                        "Graph x axis title",
                                                        "Title at bottom of graph on the X axis",
                                                        "<i>X Axis Title</i>",
                                                        G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_TITLE_Y,
                                   g_param_spec_string ("text-title-yaxis",
                                                        "Graph y axis title",
                                                        "Title on left of graph on the Y axis",
                                                        "Y Axis Title",
                                                        G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_LINE_WIDTH,
                                   g_param_spec_int  ("series-line-width",
                                                      "Series line width",
                                                      "Width of line drawn for data series",
                                                       1,10,2,             
                                                       G_PARAM_READWRITE));  
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_ELEMENTS,
                                   g_param_spec_int  ("chart-set-elements",
                                                      "Show Chart Elements",
                                                      "Enable showing these elements of the chart body",
                                                       0, GLG_RESERVED_ON, elements,             
                                                       G_PARAM_WRITABLE));  
        
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_TITLE_COLOR,
                                   g_param_spec_string ("graph-title-foreground",
                                                        "Color name",
                                                        "Main title foreground color",
                                                        "blue",
                                                        G_PARAM_WRITABLE));          
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_SCALE_COLOR,
                                   g_param_spec_string ("graph-scale-foreground",
                                                        "Color name",
                                                        "X and Y chart scale foreground font color",
                                                        "black",
                                                        G_PARAM_WRITABLE));          
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_CHART_COLOR,
                                   g_param_spec_string ("graph-chart-background",
                                                        "Color name",
                                                        "Chart inside fill color",
                                                        "light blue",
                                                        G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
                                   PROP_GRAPH_WINDOW_COLOR,
                                   g_param_spec_string ("graph-window-background",
                                                        "Color name",
                                                        "Window background fill color",
                                                        "white",
                                                        G_PARAM_WRITABLE));  
/* *** */
    g_object_class_install_property (obj_class,
  									PROP_TICK_MINOR_X,
                                   g_param_spec_int  ("range-tick-minor-x",
                                                      "x minor tick increment",
                                                      "x minor ticks on scale",
                                                       1, 100, 5,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
  									PROP_TICK_MAJOR_X,
                                   g_param_spec_int  ("range-tick-major-x",
                                                      "x major tick increment",
                                                      "x major ticks on scale",
                                                       1, 1000, 10,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
  									PROP_SCALE_MINOR_X,
                                   g_param_spec_int  ("range-scale-minor-x",
                                                      "x minor scale range",
                                                      "x minor scale range",
                                                       0, 100, 0,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
								   PROP_SCALE_MAJOR_X,
                                   g_param_spec_int  ("range-scale-major-x",
                                                      "x major scale range",
                                                      "x major scale range",
                                                       1, 1000, 100,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
  									PROP_TICK_MINOR_Y,
                                   g_param_spec_int  ("range-tick-minor-y",
                                                      "Y minor tick increment",
                                                      "Y minor ticks on scale",
                                                       1, 100, 5,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
  									PROP_TICK_MAJOR_Y,
                                   g_param_spec_int  ("range-tick-major-y",
                                                      "Y major tick increment",
                                                      "Y major ticks on scale",
                                                       1, 1000, 10,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
  									PROP_SCALE_MINOR_Y,
                                   g_param_spec_int  ("range-scale-minor-y",
                                                      "Y minor scale range",
                                                      "Y minor scale range",
                                                       0, 100, 0,             
                                                       G_PARAM_WRITABLE));  
    g_object_class_install_property (obj_class,
								   PROP_SCALE_MAJOR_Y,
                                   g_param_spec_int  ("range-scale-major-y",
                                                      "Y major scale range",
                                                      "Y major scale range",
                                                       1, 1000, 100,             
                                                       G_PARAM_WRITABLE));  
		
    g_debug ("===> glg_line_graph_class_init(exited)");
	return;
}

static void glg_line_graph_init (GlgLineGraph *graph)
{
	GlgLineGraphPrivate *priv = NULL;
		
	g_debug ("===> glg_line_graph_init(entered)");

	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

    graph->priv = priv = GLG_LINE_GRAPH_GET_PRIVATE (graph);

	gtk_widget_set_has_window (GTK_WIDGET(graph), TRUE);
	gtk_widget_set_app_paintable (GTK_WIDGET(graph), TRUE);

	priv->cb_id = GLG_PRIVATE_ID;
	priv->b_tooltip_active = FALSE;
	priv->b_mouse_onoff = FALSE;
	if (priv->series_line_width < 1) { 
		priv->series_line_width = 2;
	}

    g_debug ("===> glg_line_graph_init(exited)");
	    
	return;
}


static void glg_line_graph_realize (GtkWidget *widget)
{
  GtkAllocation allocation;
  GdkWindowAttr attributes;
  gint attributes_mask;
  GlgLineGraphPrivate *priv = NULL;
  GdkSeat *seat = NULL;

  g_debug ("===> glg_line_graph_realize(entered)");

  if (!gtk_widget_get_has_window (widget))
    {
      GTK_WIDGET_CLASS (glg_line_graph_parent_class)->realize (widget);
    }
  else
    {
      priv = GLG_LINE_GRAPH(widget)->priv;

      gtk_widget_set_realized (widget, TRUE);

      gtk_widget_get_allocation (widget, &allocation);

      attributes.window_type = GDK_WINDOW_CHILD;
      attributes.x = allocation.x;
      attributes.y = allocation.y;
      attributes.width = allocation.width;
      attributes.height = allocation.height;
      attributes.wclass = GDK_INPUT_OUTPUT;
      attributes.visual = gtk_widget_get_visual (widget);
      attributes.event_mask = gtk_widget_get_events (widget) |
                            GDK_EXPOSURE_MASK |
                        GDK_BUTTON_PRESS_MASK |
                      GDK_BUTTON_RELEASE_MASK |
                      GDK_POINTER_MOTION_MASK ;

      attributes_mask = GDK_WA_X | GDK_WA_Y | GDK_WA_VISUAL;

      priv->window = gdk_window_new (gtk_widget_get_parent_window (widget), &attributes, attributes_mask);
      gtk_widget_register_window (widget, priv->window);
      gtk_widget_set_window (widget, priv->window);

      seat = gdk_display_get_default_seat ( gtk_widget_get_display (widget) );
      priv->device_pointer = gdk_seat_get_pointer(seat);
    }

  glg_line_graph_send_configure (GLG_LINE_GRAPH(widget));

  g_debug ("===> glg_line_graph_realize(exited)");
}

static void glg_line_graph_size_allocate (GtkWidget *widget, GtkAllocation *allocation)
{
  g_debug ("===> glg_line_graph_size_allocate(entered)");

  g_return_if_fail (GLG_IS_LINE_GRAPH(widget));
  g_return_if_fail (allocation != NULL);

  gtk_widget_set_allocation (widget, allocation);

  if (gtk_widget_get_realized (widget))
    {
      if (gtk_widget_get_has_window (widget)) {
        gdk_window_move_resize (GLG_LINE_GRAPH(widget)->priv->window,
                                allocation->x, allocation->y,
                                allocation->width, allocation->height);
 	  }
      glg_line_graph_send_configure (GLG_LINE_GRAPH(widget));
    }

    g_debug ("===> glg_line_graph_size_allocate(exited)");
}

static void glg_line_graph_send_configure (GlgLineGraph *graph)
{
  GtkAllocation allocation;
  GtkWidget *widget;
  GdkEvent *event = gdk_event_new (GDK_CONFIGURE);

  g_debug ("===> glg_line_graph_send_configure(entered)");

  widget = GTK_WIDGET (graph);
  gtk_widget_get_allocation (widget, &allocation);

  event->configure.window = g_object_ref (graph->priv->window);
  event->configure.send_event = TRUE;
  event->configure.x = allocation.x;
  event->configure.y = allocation.y;
  event->configure.width = allocation.width;
  event->configure.height = allocation.height;

  gtk_widget_event (widget, event);
  gdk_event_free (event);

  g_debug ("===> glg_line_graph_send_configure(exited)");

}


/**
 * Duration
 * @param start_time  pointer to gint64 holding time since epoch in us, 0.000001
 * @param method_name gchar to name of method to log
 * @returns time since epoch in microseconds
 *
 * Note: start_time = NULL, return current time immediately
 *       method_name = NULL, skips logging of duration log message
 */
gint64 glg_duration_us(gint64 *start_time, gchar *method_name) {
	if (NULL == start_time) {
		return (g_get_real_time());
	}

	gint64 duration = (g_get_real_time() - *start_time);

	if (NULL != method_name) {
		g_debug("DURATION: %s() duration=%4.3lf ms.", method_name, (double)duration/1000);
	}
    *start_time = g_get_real_time();

	return (duration);
}

static gboolean glg_line_graph_compute_layout(GlgLineGraph *graph, GdkRectangle *allocation) {
    GlgLineGraphPrivate *priv = NULL;
    PangoFontDescription *desc = NULL;
    PangoLayout *layout = NULL;
    gint        xfactor = 0, yfactor = 0, chart_set_ranges = 0;

    g_debug ("===> glg_line_graph_compute_layout(entered)");

    g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);

    priv = graph->priv;
    g_return_val_if_fail ( priv != NULL, FALSE);

    /*
     * test to ensure chart ranges are already set */
    xfactor = MIN (priv->x_range.i_num_minor, priv->x_range.i_num_major);
    yfactor = MIN (priv->y_range.i_num_minor, priv->y_range.i_num_major);
    chart_set_ranges = MIN (xfactor, yfactor);
    g_return_val_if_fail (chart_set_ranges != 0, FALSE);

    g_debug ("===> glg_line_graph_compute_layout(new width=%d, height=%d)", allocation->width, allocation->height);

    /*
     * Compute scale: use managed or our desired user space values
     * - MUST MATCH VALUES IN <configure-event>, #compute_layout, and <draw> callbacks.
    */
    if ( (allocation->width < GLG_USER_MODEL_X) ||
         (allocation->height < GLG_USER_MODEL_Y)) {
          priv->page_box.width = GLG_USER_MODEL_X;
          priv->page_box.height = GLG_USER_MODEL_Y;
    } else {
          priv->page_box.width = allocation->width;
          priv->page_box.height = allocation->height;
    }

    /*
     * Create a PangoLayout, get the spacing of one large char to use as a standard */
    layout = gtk_widget_create_pango_layout (GTK_WIDGET(graph), NULL);
        desc = pango_font_description_from_string ("Luxi Mono 12");
        pango_layout_set_font_description (layout, desc);

        pango_layout_set_markup (layout, "<b>M</b>", -1);
        pango_layout_set_alignment (layout, PANGO_ALIGN_CENTER);
        pango_layout_get_pixel_size (layout, &xfactor, &yfactor);
        g_debug ("Alloc:factors:raw:pango_layout_get_pixel_size(width=%d, height=%d)", xfactor, yfactor);

        priv->xfactor = xfactor = ((xfactor+6)/10) * 10;
        priv->yfactor = yfactor = ((yfactor+8)/10) * 10;
        g_debug ("Alloc:factors:adj:pango_layout_get_pixel_size(width=%d, height=%d)", xfactor, yfactor);

    pango_font_description_free (desc);
    g_object_unref (layout);


    /*
     * Setup chart rectangles */
    priv->x_border = xfactor / 2;            /* def 16/2=8 edge pad */
    priv->y_border = yfactor / 4;            /* def 20/5=4 edge pad */
    
    if (priv->lgflags & GLG_TITLE_T) {        
        priv->page_title_box.x = xfactor * 6;       /* define top-left corner of textbox */
        priv->page_title_box.y = priv->y_border;
        priv->page_title_box.width  = priv->page_box.width - priv->page_title_box.x - priv->x_border;
        priv->page_title_box.height = yfactor * 2;
    }
    if (priv->lgflags & GLG_TITLE_X) {    
        priv->x_label_box.x = xfactor * 6;          /* define top-left corner of textbox */
        priv->x_label_box.y = priv->page_box.height - yfactor - priv->y_border - priv->x_border;
        priv->x_label_box.width  = priv->page_box.width - priv->x_label_box.x - priv->x_border;
        priv->x_label_box.height = yfactor + priv->y_border;
    }
    if (priv->lgflags & GLG_TITLE_Y) {              /* define bottom left corner */
        priv->y_label_box.x = priv->x_border;
        priv->y_label_box.y = priv->page_box.height - (yfactor * 3);
        priv->y_label_box.width  = xfactor * 3;
        priv->y_label_box.height = priv->y_label_box.y - (yfactor * 3);
    }
    if (priv->lgflags & GLG_TOOLTIP) {        
        priv->tooltip_box.x = priv->y_label_box.width + priv->y_label_box.x + (xfactor * 2) + priv->x_border;       /* define top-left corner of textbox */
        priv->tooltip_box.y = priv->y_border;
        priv->tooltip_box.width = priv->page_box.width - priv->tooltip_box.x - xfactor;
        priv->tooltip_box.height = (yfactor * 2) + priv->y_border;
    }

    /*
     * This is for the main chart area or plot box 
     * -- this calc is for the maximum available area 
     */
    priv->plot_box.x = priv->y_label_box.width + priv->y_label_box.x + (xfactor * 3);
    priv->plot_box.y = priv->page_title_box.height + priv->page_title_box.y + priv->y_border;
    priv->plot_box.width  = (gint) priv->page_box.width - priv->plot_box.x - xfactor;
    priv->plot_box.height = (gint) priv->page_box.height - priv->plot_box.y  - priv->x_label_box.height - yfactor;
   
    g_debug ("Alloc:Max.Avail: plot_box.width=%d, plot_box.height=%d", priv->plot_box.width, priv->plot_box.height);

    /* 
     * reposition the box according to scale-able increments 
     * -- this calc is to align to scaling requirements 
     */
    xfactor = priv->plot_box.width;
    yfactor = priv->plot_box.height;
    priv->plot_box.width  = ( (gint)(priv->plot_box.width  / priv->x_range.i_num_minor) * priv->x_range.i_num_minor);  
    priv->plot_box.height = ( (gint)(priv->plot_box.height / priv->y_range.i_num_minor) * priv->y_range.i_num_minor);

    /*
     * Distribute the difference toward the bottom right
     */
    xfactor -= priv->plot_box.width;
    yfactor -= priv->plot_box.height;
    priv->plot_box.x += (gint)(xfactor * 0.80);
    priv->plot_box.y += yfactor;    
    priv->tooltip_box.x = priv->page_title_box.x = priv->x_label_box.x = priv->plot_box.x;
    priv->tooltip_box.width = priv->x_label_box.width = priv->page_title_box.width = priv->plot_box.width;
    priv->y_label_box.y = priv->plot_box.y + priv->plot_box.height;

    /*  
     * Determine the pixel increment of the grid lines 
     */
    priv->y_range.i_minor_inc = priv->plot_box.height / priv->y_range.i_num_minor;
    priv->y_range.i_major_inc = priv->plot_box.height / priv->y_range.i_num_major;
    priv->x_range.i_minor_inc = priv->plot_box.width / priv->x_range.i_num_minor;
    priv->x_range.i_major_inc = priv->plot_box.width / priv->x_range.i_num_major;

    g_debug ("Alloc:Chart:Incs:    x_minor=%d, x_major=%d, y_minor=%d, y_major=%d, plot_box.x=%d, plot_box.y=%d, plot_box.width=%d, plot_box.height=%d",
          priv->x_range.i_minor_inc,
          priv->x_range.i_major_inc,
          priv->y_range.i_minor_inc,
          priv->y_range.i_major_inc,
          priv->plot_box.x,
          priv->plot_box.y,
          priv->plot_box.width,
          priv->plot_box.height);

    g_debug ("Alloc:Chart:Nums:    x_num_minor=%d, x_num_major=%d, y_num_minor=%d, y_num_major=%d",
          priv->x_range.i_num_minor,
          priv->x_range.i_num_major,
          priv->y_range.i_num_minor,
          priv->y_range.i_num_major);

    g_debug ("Alloc:Chart:Plot:    x=%d, y=%d, width=%d, height=%d",
            priv->plot_box.x,
            priv->plot_box.y,
            priv->plot_box.width,
            priv->plot_box.height);

    g_debug ("Alloc:Chart:Title:   x=%d, y=%d, width=%d, height=%d",
            priv->page_title_box.x,
            priv->page_title_box.y,
            priv->page_title_box.width,
            priv->page_title_box.height);
    g_debug ("Alloc:Chart:yLabel:  x=%d, y=%d, width=%d, height=%d",
            priv->y_label_box.x,
            priv->y_label_box.y,
            priv->y_label_box.width,
            priv->y_label_box.height);
    g_debug ("Alloc:Chart:xLabel:  x=%d, y=%d, width=%d, height=%d",
            priv->x_label_box.x,
            priv->x_label_box.y,
            priv->x_label_box.width,
            priv->x_label_box.height);
    g_debug ("Alloc:Chart:Tooltip: x=%d, y=%d, width=%d, height=%d",
            priv->tooltip_box.x,
            priv->tooltip_box.y,
            priv->tooltip_box.width,
            priv->tooltip_box.height);

    g_debug ("===> glg_line_graph_compute_layout(exited)");
    
    return (TRUE);
}

static gboolean glg_line_graph_configure_event (GtkWidget *widget, GdkEventConfigure *event)
{
	GlgLineGraphPrivate *priv;
    GlgLineGraph   *graph = GLG_LINE_GRAPH(widget);
    GtkAllocation allocation;
	cairo_t *cr;
    cairo_status_t status;
    gint width = 0, height = 0;


	g_debug ("===> glg_line_graph_configure_event(entered)");

	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);
	g_return_val_if_fail ( event->type == GDK_CONFIGURE, FALSE);

	priv = graph->priv;
	g_return_val_if_fail ( priv != NULL, FALSE);

	/*
	 *  Compute the graph boxes sizing */
     allocation.x = event->x;
     allocation.y = event->y;
     allocation.width = event->width;
     allocation.height = event->height;
    glg_line_graph_compute_layout(graph, &allocation);

    /*
     * Create an Image Source for off-line drawing */
    if (priv->surface) {
        cairo_surface_destroy (priv->surface);
    }

    width = event->width;
    height = event->height;

    /*
     * Compute scale: use managed or our desired user space values
     * - MUST MATCH VALUES IN <configure-event>, #compute_layout, and <draw> callbacks.
    */
	if ((width < GLG_USER_MODEL_X) ||
       (height < GLG_USER_MODEL_Y)) {
			width = GLG_USER_MODEL_X;
			height = GLG_USER_MODEL_Y;
	}


    priv->surface = gdk_window_create_similar_surface (priv->window , /* TODO: Poor Performance: also CAIRO_CONTENT_COLOR_ALPHA */
                                                   CAIRO_CONTENT_COLOR_ALPHA,
    											       width, height);
	status = cairo_surface_status(priv->surface);
	if (status != CAIRO_STATUS_SUCCESS) {
		g_message ("GLG-Configure-Event:#cairo_image_surface_create:status %d=%s", status, cairo_status_to_string(status) );
	} else {
		cr = cairo_create (priv->surface);
			cairo_set_source_rgba (cr, priv->window_color.red, priv->window_color.green, priv->window_color.blue, 0.8);
			cairo_paint (cr);
		cairo_destroy (cr);

	    glg_line_graph_draw_graph (GTK_WIDGET(widget));
	}

    g_debug ("===> glg_line_graph_configure_event(exited)");
    
    return (TRUE);
}

/*
 * This routine is the master paint routine for the line graph
 * it sets up the dimensions of the space and call DRAW to get the graph done
*/
static gboolean glg_line_graph_master_draw (GtkWidget *graph, cairo_t *cr)
{
	GlgLineGraphPrivate *priv;
    gint        xfactor = 0, yfactor = 0, chart_set_ranges = 0;
    GtkWidget   *widget = GTK_WIDGET(graph);
    GtkAllocation allocation;
    GdkRectangle dirtyRect;
    gint64 start_time = glg_duration_us(NULL, NULL);


	g_debug ("===> glg_line_graph_master_draw(entered)");

	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);
	
	priv = GLG_LINE_GRAPH(graph)->priv;
	g_return_val_if_fail ( priv != NULL, FALSE);	

	gtk_widget_get_allocation(widget, &allocation);
	gdk_cairo_get_clip_rectangle (cr, &dirtyRect);

   	g_debug ("glg_line_graph_master_draw(Allocation ==> width=%d, height=%d,  Dirty Rect ==> x=%d, y=%d, width=%d, height=%d )",
    	                 allocation.width, allocation.height,
    	                 dirtyRect.x, dirtyRect.y, dirtyRect.width, dirtyRect.height);

	/* 
	 * test to ensure chart ranges are already set */
	xfactor = MIN (priv->x_range.i_num_minor, priv->x_range.i_num_major);
	yfactor = MIN (priv->y_range.i_num_minor, priv->y_range.i_num_major);
	chart_set_ranges = MIN (xfactor, yfactor);    
	g_return_val_if_fail (chart_set_ranges != 0, FALSE);

    /*
     * Compute scale: use managed or our desired user space values
     * - MUST MATCH VALUES IN <configure-event>, #compute_layout, and <draw> callbacks.
    */
	if ((allocation.width < GLG_USER_MODEL_X) ||
       (allocation.height < GLG_USER_MODEL_Y)) {
       	
	    cairo_scale (cr, (gdouble)allocation.width/GLG_USER_MODEL_X,
		  		 		 (gdouble)allocation.height/GLG_USER_MODEL_Y);
		  		 		 
		g_debug ("glg_line_graph_master_draw#cairo_scale( x=%3.3f, y=%3.3f)",
								 (gdouble)allocation.width/GLG_USER_MODEL_X,
							     (gdouble)allocation.height/GLG_USER_MODEL_Y);
    }

	/*
	 * set source after determining if scaling is required */
    cairo_set_source_surface (cr, priv->surface, 0, 0);
    cairo_paint (cr);

   	glg_duration_us(&start_time, "glg_line_graph_master_draw#TOTAL-TIME");
    g_debug ("glg_line_graph_master_draw(exited)");

	return TRUE;
}

/**
 * glg_line_graph_redraw:
 * @graph: pointer to a #GlgLineGraph widget
 * 
 * Updates the current graph showing new changes.
*/
extern void glg_line_graph_redraw (GlgLineGraph *graph)
{
	GtkAllocation allocation;
	
	g_debug ("===> glg_line_graph_redraw(entered)");

	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

    /* redraw the window completely by exposing it */
    glg_line_graph_draw_graph (GTK_WIDGET(graph));
    	gtk_widget_get_allocation(GTK_WIDGET (graph), &allocation);
    	gtk_widget_queue_draw_area (GTK_WIDGET (graph), allocation.x, allocation.y, allocation.width, allocation.height);

	g_debug ("===> glg_line_graph_redraw(exited)");
}

/**
 * glg_line_graph_new:
 * @first_property_name: NULL or gchar pointer to first property name to set, next param must be its value
 * @...: va_list null terminated list of additional property-name/property-value pairs
 *
 * Creates a new line graph widget
 * - optionally accepts 'property-name, property-values',...,NULL pairs
 * Returns:  #GlgLineGraph widget
*/
extern GlgLineGraph * glg_line_graph_new (const gchar *first_property_name, ...)
{
	GlgLineGraph *graph = NULL;
    va_list var_args;
	
	g_debug ("===> glg_line_graph_new(entered)");

    g_return_val_if_fail (first_property_name != NULL, g_object_new(GLG_TYPE_LINE_GRAPH, NULL));

    va_start (var_args, first_property_name);
    graph = (GlgLineGraph *)g_object_new_valist (GLG_TYPE_LINE_GRAPH, first_property_name, var_args);
    va_end (var_args);

	g_debug ("===> glg_line_graph_new(exited)");
    return graph;		
}

/**
 * glg_line_graph_chart_set_x_ranges:
 * @graph: pointer to a #GlgLineGraph widget
 * @x_tick_minor: number of minor divisions for x scale
 * @x_tick_major: number of major divisions for x scale
 * @x_scale_min:  minimum x scale value or starting point
 * @x_scale_max:  maximum x scale value or ending point
 * 
 * Sets the X ticks and scales for the graph grid area.
 */
extern void glg_line_graph_chart_set_x_ranges (GlgLineGraph *graph,
                                 gint x_tick_minor, gint x_tick_major,
                                 gint x_scale_min,  gint x_scale_max)
{
    GlgLineGraphPrivate *priv;
    gint xfactor = 0;

    g_debug ("===> glg_line_graph_chart_set_x_ranges()");

    g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

    priv = graph->priv;

    xfactor = MIN( x_tick_minor, x_tick_major );
    g_return_if_fail ( xfactor != 0);           /* test for contextually vaild input */

    if ( priv->x_range.i_max_scale ) {
        g_message ("Set Ranges Failed: Cannot set ranges more than once, range already set!");
        return;
    }

    priv->x_range.i_inc_minor_scale_by = x_tick_minor;  /* minimum scale value - ex:   1 */
    priv->x_range.i_inc_major_scale_by = x_tick_major;  /* minimum scale value - ex:   1 */

    priv->x_range.i_min_scale = x_scale_min;   /* minimum scale value - ex:   0 */
    priv->x_range.i_max_scale = x_scale_max;   /* maximum scale value - ex: 100 */
    priv->x_range.i_num_minor = x_scale_max / x_tick_minor;   /* number of minor points */
    priv->x_range.i_num_major = x_scale_max / x_tick_major;   /* number of major points */

    return;
}

/**
 * glg_line_graph_chart_set_y_ranges:
 * @graph: pointer to a #GlgLineGraph widget
 * @y_tick_minor: number of minor divisions for y scale
 * @y_tick_major: number of major divisions for y scale
 * @y_scale_min:  minimum y scale value or starting point
 * @y_scale_max:  maximum y scale value or ending point
 *
 * Sets the Y ticks and scales for the graph grid area.
 */
extern void glg_line_graph_chart_set_y_ranges (GlgLineGraph *graph,
                                 gint y_tick_minor, gint y_tick_major,
                                 gint y_scale_min,  gint y_scale_max )
{
    GlgLineGraphPrivate *priv;
    gint yfactor = 0;

    g_debug ("===> glg_line_graph_chart_set_y_ranges()");

    g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

    priv = graph->priv;

    yfactor = MIN( y_tick_minor, y_tick_major );
    g_return_if_fail ( yfactor != 0); /* test for contextually vaild input */

    if ( priv->y_range.i_max_scale ) {
        g_message ("Set Y Ranges Failed: Cannot set ranges more than once, range already set!");
        return;
    }

    priv->y_range.i_inc_minor_scale_by = y_tick_minor;  /* minimum scale value - ex:   1 */
    priv->y_range.i_inc_major_scale_by = y_tick_major;  /* minimum scale value - ex:   0 */

    priv->y_range.i_min_scale = y_scale_min;   /* minimum scale value - ex:   0 */
    priv->y_range.i_max_scale = y_scale_max;   /* maximum scale value - ex: 100 */
    priv->y_range.i_num_minor = y_scale_max / y_tick_minor;   /* number of minor points */
    priv->y_range.i_num_major = y_scale_max / y_tick_major;   /* number of major points */

    return;
}


/**
 * glg_line_graph_chart_set_ranges:
 * @graph: pointer to a #GlgLineGraph widget
 * @x_tick_minor: number of minor divisions for x scale
 * @x_tick_major: number of major divisions for x scale
 * @x_scale_min:  minimum x scale value or starting point
 * @x_scale_max:  maximum x scale value or ending point
 * @y_tick_minor: number of minor divisions for y scale
 * @y_tick_major: number of major divisions for y scale
 * @y_scale_min:  minimum y scale value or starting point
 * @y_scale_max:  maximum y scale value or ending point
 *
 * Sets the X and Y ticks and scales for the graph grid area.
 */
extern void glg_line_graph_chart_set_ranges (GlgLineGraph *graph,
                                 gint x_tick_minor, gint x_tick_major,
                                 gint x_scale_min,  gint x_scale_max,
                                 gint y_tick_minor, gint y_tick_major, 
                                 gint y_scale_min,  gint y_scale_max)
{
	GlgLineGraphPrivate *priv;
	gint xfactor = 0, yfactor = 0, chart_set_ranges = 0;

	g_debug ("===> glg_line_graph_chart_set_ranges()");

	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));
	
	priv = graph->priv;
	
	xfactor = MIN( x_tick_minor, x_tick_major );
	yfactor = MIN( y_tick_minor, y_tick_major );
	
	chart_set_ranges = MIN( xfactor, yfactor);    /* test for contextually vaild input */			
	g_return_if_fail ( chart_set_ranges != 0);
	
	if ( priv->x_range.i_max_scale ) {
		g_message ("Set Ranges Failed: Cannot set ranges more than once, range already set!");
		return;
	}

    priv->x_range.i_inc_minor_scale_by = x_tick_minor;  /* minimum scale value - ex:   1 */
    priv->x_range.i_inc_major_scale_by = x_tick_major;  /* minimum scale value - ex:   1 */

    priv->x_range.i_min_scale = x_scale_min;   /* minimum scale value - ex:   0 */
    priv->x_range.i_max_scale = x_scale_max;   /* maximum scale value - ex: 100 */
    priv->x_range.i_num_minor = x_scale_max / x_tick_minor;   /* number of minor points */
    priv->x_range.i_num_major = x_scale_max / x_tick_major;   /* number of major points */

    priv->y_range.i_inc_minor_scale_by = y_tick_minor;  /* minimum scale value - ex:   1 */
    priv->y_range.i_inc_major_scale_by = y_tick_major;  /* minimum scale value - ex:   0 */

    priv->y_range.i_min_scale = y_scale_min;   /* minimum scale value - ex:   0 */
    priv->y_range.i_max_scale = y_scale_max;   /* maximum scale value - ex: 100 */
    priv->y_range.i_num_minor = y_scale_max / y_tick_minor;   /* number of minor points */
    priv->y_range.i_num_major = y_scale_max / y_tick_major;   /* number of major points */

    return;
}

/**
 * glg_line_graph_chart_set_color:
 * @graph: pointer to a #GlgLineGraph widget
 * @element: #GLGElementID of the Window Element 
 * @pch_color: gchar* string with name of color.
 * 
 * Copy (without freeing) pch_color into place to be used as the graph element color. 
 * #GLG_SCALE  - x/y integer legends color {default black}
 * #GLG_TITLE  - Main graph title. {default light blue}
 * #GLG_WINDOW - Graph window background color, and grid foreground. {default white}
 * #GLG_CHART  - Graph plot area background. {default light blue}
 *
 * Returns: gboolean TRUE if copied, FALSE if copy failed.
 */
extern gboolean glg_line_graph_chart_set_color (GlgLineGraph *graph, GLGElementID element, const gchar * pch_color)
{
    GlgLineGraphPrivate *priv;
    gboolean   rc = TRUE;


	g_debug ("===> glg_line_graph_chart_set_color(entered)");

	g_return_val_if_fail (GLG_IS_LINE_GRAPH(graph), FALSE);
    g_return_val_if_fail (pch_color != NULL, FALSE);        
	
	priv = graph->priv;


    switch ( element ) {
        case GLG_SCALE:
             g_utf8_strncpy (priv->ch_color_scale_fg, pch_color, sizeof (priv->ch_color_scale_fg));
             gdk_rgba_parse (&priv->scale_color, priv->ch_color_scale_fg);
             break;        
        case GLG_TITLE:
            g_utf8_strncpy (priv->ch_color_title_fg, pch_color, sizeof (priv->ch_color_title_fg));
             gdk_rgba_parse (&priv->title_color, priv->ch_color_title_fg);
             break;        
        case GLG_WINDOW:
            g_utf8_strncpy (priv->ch_color_window_bg, pch_color, sizeof (priv->ch_color_window_bg));
             gdk_rgba_parse (&priv->window_color, priv->ch_color_window_bg);
             break;
        case GLG_CHART:
            g_utf8_strncpy (priv->ch_color_chart_bg, pch_color, sizeof (priv->ch_color_chart_bg));
             gdk_rgba_parse (&priv->chart_color, priv->ch_color_chart_bg);
             break;        
        default:
             g_message ("glg_line_graph_chart_set_color(): Invalid Element ID");
             rc = FALSE;
    }
    
	g_debug ("===> glg_line_graph_chart_set_color(exited)");
   return rc;
}

/**
 * glg_line_graph_chart_set_text:
 * @graph: pointer to a #GlgLineGraph widget
 * @element: #GLGElementID of the Title 
 * @pch_text: gchar* string to set in title.
 * 
 * Copy (without freeing) pch_text into place to be used as the graph title. PangoMarkup is
 * also supported.
 *  
 * #GLG_TITLE_X - Bottom x axis title
 * #GLG_TITLE_Y - Left vertical y axis title
 * #GLG_TITLE_T - Top and considered main title on x axis
 * 
 * Returns: gboolean TRUE if copied, FALSE if copy failed.
 */
extern gboolean glg_line_graph_chart_set_text (GlgLineGraph *graph, GLGElementID element, const gchar *pch_text)
{
	GlgLineGraphPrivate *priv;
    gchar 	   *pch = NULL;
    gboolean   rc = TRUE;

	g_debug ("===> glg_line_graph_chart_set_text(entered)");

	g_return_val_if_fail (GLG_IS_LINE_GRAPH(graph), FALSE);
    g_return_val_if_fail (pch_text != NULL, FALSE);        
	
	priv = graph->priv;

    pch = g_strdup (pch_text);
    
    switch ( element ) {
        case GLG_TITLE_X:
             if (priv->x_label_text != NULL) {
                 g_free (priv->x_label_text);
             }
             priv->x_label_text = pch;
             break;        
        case GLG_TITLE_Y:
             if (priv->y_label_text != NULL) {
                 g_free (priv->y_label_text);
             }
             priv->y_label_text = pch;
             break;        
        case GLG_TITLE_T:
             if (priv->page_title_text != NULL) {
                 g_free (priv->page_title_text);
             }
             priv->page_title_text = pch;
             break;
        case GLG_TOOLTIP:
             g_snprintf (priv->ch_tooltip_text, sizeof (priv->ch_tooltip_text), "%s", pch);
             g_free ( pch );
             break;        
        default:
             g_message ("glg_line_graph_chart_set_text(): Invalid Element ID");
             g_free ( pch );             
             rc = FALSE;
    }
    
	g_debug ("===> glg_line_graph_chart_set_text(exited)");
  return rc;
}

static void glg_line_graph_draw_graph (GtkWidget *graph)
{
	GlgLineGraphPrivate *priv;
	GLGElementID element = 0;
    gint64 start_time = glg_duration_us(NULL, NULL);
    gint64 duration = glg_duration_us(NULL, NULL);

	g_debug ("===> glg_line_graph_draw_graph(entered)");

	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

	priv = GLG_LINE_GRAPH(graph)->priv;

    /* Paint to the surface, where we store our state */
	  priv->cr = cairo_create (priv->surface);
	  	cairo_set_source_rgba (priv->cr, 1, 1, 1, 0.9);
	  	cairo_paint (priv->cr);
	
    /* 
     * draw plot area */
    cairo_set_source_rgba (priv->cr, (gdouble)priv->chart_color.red,
    								 (gdouble)priv->chart_color.green,
    								 (gdouble)priv->chart_color.blue, 0.8);
	cairo_rectangle (priv->cr, priv->plot_box.x, priv->plot_box.y, priv->plot_box.width, priv->plot_box.height);
	cairo_fill_preserve (priv->cr);
    cairo_set_source_rgba (priv->cr, 0., 0., 0., 0.8);   /* black */
    cairo_stroke (priv->cr);
    glg_duration_us(&start_time, "glg_line_graph_draw_graph#PlotArea");

		g_debug ("Chart.Surface: pg.Width=%d, pg.Height=%d, Plot Area x=%d y=%d width=%d, height=%d",
			  priv->page_box.width, priv->page_box.height,priv->plot_box.x, priv->plot_box.y, priv->plot_box.width, priv->plot_box.height);

		/*
		 * draw titles
		 */
		element = priv->lgflags;
		if ( element & GLG_TITLE_T) {
			glg_line_graph_draw_text_horizontal (GLG_LINE_GRAPH(graph), priv->page_title_text, &priv->page_title_box);
			glg_duration_us(&start_time, "glg_line_graph_draw_graph#Top-Title");
		}
		if ( element & GLG_TITLE_X) {
			glg_line_graph_draw_text_horizontal (GLG_LINE_GRAPH(graph), priv->x_label_text, &priv->x_label_box);
			glg_duration_us(&start_time, "glg_line_graph_draw_graph#X-Title");
		}
		if ( element & GLG_TITLE_Y) {
			glg_line_graph_draw_text_vertical (GLG_LINE_GRAPH(graph), priv->y_label_text, &priv->y_label_box);
			glg_duration_us(&start_time, "glg_line_graph_draw_graph#Y-Title");
		}

		if ( ( element & GLG_GRID_LINES ) |
			 ( element & GLG_GRID_MINOR_X ) |
			 ( element & GLG_GRID_MAJOR_X ) |
			 ( element & GLG_GRID_MINOR_Y ) |
			 ( element & GLG_GRID_MAJOR_Y) ) {
			 glg_line_graph_draw_grid_lines (GLG_LINE_GRAPH(graph));
			 glg_duration_us(&start_time, "glg_line_graph_draw_graph#GridLines");
		}
		if ( element & GLG_GRID_LABELS_X) {
			 glg_line_graph_draw_x_grid_labels (GLG_LINE_GRAPH(graph));
			 glg_duration_us(&start_time, "glg_line_graph_draw_graph#X-Labels");
		}
		if ( element & GLG_GRID_LABELS_Y) {
			 glg_line_graph_draw_y_grid_labels (GLG_LINE_GRAPH(graph));
			 glg_duration_us(&start_time, "glg_line_graph_draw_graph#Y-Labels");
		}

		glg_line_graph_data_series_draw_all (GLG_LINE_GRAPH(graph), FALSE);
		glg_duration_us(&start_time, "glg_line_graph_draw_graph#Series-All");

		if ( element & GLG_TOOLTIP) {
			 glg_line_graph_draw_tooltip (GLG_LINE_GRAPH(graph));
			 glg_duration_us(&start_time, "glg_line_graph_draw_graph#Tooltip");
		}

    cairo_destroy (priv->cr);
    priv->cr = NULL;

    g_debug ("===> glg_line_graph_draw_graph(exited)");
    glg_duration_us(&duration, "glg_line_graph_draw_graph#TOTAL-TIME");
	
	return;	
}


/*
 * Draws a label text on the X axis
 * sets the width, height values of the input rectangle to the size of textbox
 * returns the width of the text area, or -1 on error
*/
static gint glg_line_graph_draw_text_horizontal (GlgLineGraph *graph, gchar * pch_text, cairo_rectangle_int_t * rect)
{
    GlgLineGraphPrivate *priv;
    PangoLayout *layout = NULL;
    gint        x_pos = 0, y_pos = 0, width = 0, height = 0;

	g_debug ("===> glg_line_graph_draw_text_horizontal()");

    g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);    
    if (pch_text == NULL ) { return -1; }    
    g_return_val_if_fail (rect != NULL, -1);
    priv = graph->priv;
    g_return_val_if_fail (priv->cr != NULL, -1);


    /*
     * Get pixel size in user space coordinates */

  	layout = pango_cairo_create_layout (priv->cr);
	pango_layout_set_markup (layout, pch_text, -1);
    pango_layout_set_alignment (layout, PANGO_ALIGN_CENTER);
	pango_cairo_update_layout (priv->cr, layout);
    pango_layout_get_pixel_size (layout, &width, &height);
    if (width > rect->width ) {
 	    x_pos = (gint)(priv->page_box.width - width) / 2;
    } else {
    	x_pos = rect->x + (gint)((rect->width - width) /2);
    }
    if (height > rect->height) {
 	    y_pos = MAX((rect->y - (gint)(height - rect->height)), 0);
 	     
	} else {
	    y_pos = rect->y + (gint)((rect->height - height) * 0.80);
	}

	g_debug ("Horiz.TextBox:Page cx=%d, cy=%d", 
				  priv->page_box.width, priv->page_box.height);
	g_debug ("Horiz.TextBox:Orig: x=%d, y=%d, cx=%d, cy=%d", 
				  rect->x, rect->y, rect->width, rect->height);
	g_debug ("Horiz.TextBox:Calc x_pos=%d, y_pos=%d,  cx=%d, cy=%d", 
				  x_pos, y_pos, width, height);

	/* title_gc */	
    cairo_set_source_rgb (priv->cr, (gdouble)priv->title_color.red,
    								(gdouble)priv->title_color.green,
    								(gdouble)priv->title_color.blue);
	cairo_move_to (priv->cr, x_pos, y_pos); 
	pango_cairo_show_layout (priv->cr, layout);

  	g_object_unref (layout);

    return rect->width;
}

/*
 * Draws a label text on the Y axis
 * sets the width, height values of the input rectangle to the size of textbox
 * returns the height of the text area, or -1 on error
*/
static gint glg_line_graph_draw_text_vertical (GlgLineGraph *graph, gchar *pch_text, cairo_rectangle_int_t *rect)
{
    GlgLineGraphPrivate *priv = NULL;
    PangoLayout *layout = NULL;
    gint        y_pos = 0;

	g_debug ("===> glg_line_graph_draw_text_vertical()");

    g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);
    if (pch_text == NULL) { return 1; }
    g_return_val_if_fail (rect != NULL, -1);

    priv = graph->priv;

	cairo_save (priv->cr);

		/*
		 * Get pixel size in user space coordinates */
		layout = pango_cairo_create_layout (priv->cr);
		pango_layout_set_markup (layout, pch_text, -1);
		pango_layout_set_alignment (layout, PANGO_ALIGN_CENTER);

		pango_layout_get_pixel_size (layout, &rect->width, &rect->height);
		if (priv->plot_box.height > rect->width ) {
			y_pos = rect->y - ((priv->plot_box.height - rect->width) / 2);
		} else {
			y_pos = priv->page_box.height - ((priv->page_box.height - rect->width) / 2);
		}

		g_debug ("Vert:TextBox: y_pos=%d,  x=%d, y=%d, cx=%d, cy=%d",
						y_pos, rect->x, rect->y, rect->width, rect->height);

		/* title_gc */
		cairo_set_source_rgb (priv->cr, (gdouble)priv->title_color.red,
										(gdouble)priv->title_color.green,
										(gdouble)priv->title_color.blue);
		cairo_move_to (priv->cr, rect->x, y_pos);

		cairo_rotate(priv->cr, G_PI / -2.);

		pango_cairo_update_layout (priv->cr, layout);
		pango_cairo_show_layout (priv->cr, layout);

		g_object_unref (layout);

  	cairo_restore (priv->cr);

	g_debug ("Vert.TextBox: y_pos=%d,  x=%d, y=%d, cx=%d, cy=%d",
				  y_pos, rect->x, rect->y, rect->width, rect->height);

    return rect->height;
}


/*
 * Draws the minor and major grid lines inside the current plot_area
 * returns -1 on error, or TRUE;
*/
static gint glg_line_graph_draw_grid_lines (GlgLineGraph *graph)
{
	GlgLineGraphPrivate *priv;
    gint        y_minor_inc = 0, y_pos = 0, y_index = 0;
    gint        y_major_inc = 0;
    gint        x_minor_inc = 0, x_pos = 0, x_index = 0;
    gint        x_major_inc = 0;
    gint        count_major = 0, count_minor = 0;


    g_debug ("===> glg_line_graph_draw_grid_lines()");

	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);

	priv = graph->priv;

   	cairo_set_source_rgba (priv->cr, (gdouble)priv->window_color.red,
    								 (gdouble)priv->window_color.green,
    								 (gdouble)priv->window_color.blue, 0.6);
 
    count_major = priv->y_range.i_num_major - 1;
    count_minor = priv->y_range.i_num_minor - 1;
    y_minor_inc = priv->y_range.i_minor_inc;
    y_major_inc = priv->y_range.i_major_inc;

	g_debug
            ("Draw.Y-GridLines: count_major=%d, count_minor=%d, y_minor_inc=%d, y_major_inc=%d",
             count_major, count_minor, y_minor_inc, y_major_inc);

    x_pos = priv->plot_box.width;
    y_pos = priv->plot_box.y;    
    if (priv->lgflags & GLG_GRID_MINOR_Y) {    
  		cairo_set_line_width (priv->cr, 1.0);
	    for (y_index = 0; y_index < count_minor; y_index++)
    	{
    		cairo_move_to (priv->cr, priv->plot_box.x+1, y_pos + (y_minor_inc * (y_index + 1)) );
    		cairo_rel_line_to (priv->cr, x_pos - 2, 0 );
    	}
  		cairo_stroke (priv->cr);
    }


    x_pos = priv->plot_box.width;
    y_pos = priv->plot_box.y;
    if (priv->lgflags & GLG_GRID_MAJOR_Y) {
    	cairo_set_line_width (priv->cr, 2.0);
    	for (y_index = 0; y_index < count_major; y_index++)
    	{
	    	cairo_move_to (priv->cr, priv->plot_box.x, y_pos + (y_major_inc * (y_index + 1)) );
    		cairo_rel_line_to (priv->cr, x_pos - 2, 0 );
    	}
    	cairo_stroke (priv->cr);
    	cairo_set_line_width (priv->cr, 1.0);
    }


    count_major = priv->x_range.i_num_major -1;
    count_minor = priv->x_range.i_num_minor -1;
    x_minor_inc = priv->x_range.i_minor_inc;
    x_major_inc = priv->x_range.i_major_inc;

    g_debug ("Draw.X-GridLines: count_major=%d, count_minor=%d, x_minor_inc=%d, x_major_inc=%d",
             count_major, count_minor, x_minor_inc, x_major_inc);

    x_pos = priv->plot_box.x;
    y_pos = priv->plot_box.height;
    if (priv->lgflags & GLG_GRID_MINOR_X) {
  		cairo_set_line_width (priv->cr, 1.0);
    	for (x_index = 0; x_index < count_minor; x_index++)
    	{
    		cairo_move_to (priv->cr, priv->plot_box.x + (x_minor_inc * (x_index + 1)), priv->plot_box.y +1 ); 
    		cairo_line_to (priv->cr, priv->plot_box.x + (x_minor_inc * (x_index + 1)), priv->plot_box.y + y_pos -1 );
    	}
  		cairo_stroke (priv->cr);
    }

    x_pos = priv->plot_box.x;
    y_pos = priv->plot_box.height;
    if (priv->lgflags & GLG_GRID_MAJOR_X) {
		cairo_set_line_width (priv->cr, 2.0);
	    for (x_index = 0; x_index < count_major; x_index++)
	    {
	    	cairo_move_to (priv->cr, priv->plot_box.x + (x_major_inc * (x_index + 1)), priv->plot_box.y +1 ); 
    		cairo_line_to (priv->cr, priv->plot_box.x + (x_major_inc * (x_index + 1)), priv->plot_box.y  + y_pos );
	    }
    	cairo_stroke(priv->cr);
		cairo_set_line_width (priv->cr, 1.0);        
    }
    
    return TRUE;
}

/*
 * Draw the chart x scale legend
*/
static void glg_line_graph_draw_x_grid_labels (GlgLineGraph *graph)
{
	GlgLineGraphPrivate *priv;
    gchar       ch_grid_label[GLG_MAX_BUFFER];
    gchar       ch_work[GLG_MAX_BUFFER];
    PangoLayout *layout = NULL;
    PangoTabArray *p_tabs = NULL;
    gint        x_adj = 0, x1_adj = 0, width = 0, height = 0, h_index = 0, x_scale = 0,
                cx = 0, cy = 0;

    g_debug ("===> glg_line_graph_draw_x_grid_labels()");
	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

	priv = graph->priv;
    
    g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "<span font_desc=\"Monospace 8\">%d</span>",
                priv->x_range.i_max_scale);
    layout = pango_cairo_create_layout (priv->cr);
    	pango_layout_set_markup (layout, ch_grid_label, -1);
    	pango_layout_get_pixel_size (layout, &width, &height);
    x_adj = width / 2;
    x1_adj = width / 4;

	g_debug ("Scale:Labels:X small font sizes cx=%d, cy=%d", width, height);

    g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "<span font_desc=\"Monospace 8\">%s", "0");
    for (h_index = priv->x_range.i_inc_major_scale_by;
         h_index <= priv->x_range.i_max_scale;
         h_index += priv->x_range.i_inc_major_scale_by)
    {
        g_strlcpy (ch_work, ch_grid_label, GLG_MAX_BUFFER);
        g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "%s\t%d", ch_work, h_index);
        if (h_index < 10)
        {
            x_scale++;
        }
    }
    g_strlcpy (ch_work, ch_grid_label, GLG_MAX_BUFFER);
    g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "%s</span>", ch_work);

    pango_layout_set_markup (layout, ch_grid_label, -1);

    p_tabs = pango_tab_array_new (priv->x_range.i_num_major, TRUE);
    for (h_index = 0; h_index <= priv->x_range.i_num_major; h_index++)
    {
        gint        xbase = 0;

        if (h_index > x_scale)
        {
            xbase = (h_index * priv->x_range.i_major_inc);
        }
        else
        {
            xbase = (h_index * priv->x_range.i_major_inc) + x1_adj;
        }
        if (h_index == 0)
        {
            xbase = priv->x_range.i_major_inc + x1_adj;
        }
        pango_tab_array_set_tab (p_tabs, h_index, PANGO_TAB_LEFT, xbase);
    }
    pango_layout_set_tabs (layout, p_tabs);

    pango_cairo_update_layout (priv->cr, layout);

    pango_layout_get_pixel_size (layout, &cx, &cy);

    g_debug ("Scale:Labels:X plot_box.cx=%d, layout.cx=%d, layout.cy=%d", priv->plot_box.width, cx, cy);

    if ( priv->page_box.width > cx ) {
	     cairo_set_source_rgba (priv->cr, (gdouble)priv->scale_color.red,
    								 (gdouble)priv->scale_color.green,
    								 (gdouble)priv->scale_color.blue, 0.6);
		 
         cairo_move_to (priv->cr, priv->plot_box.x - x_adj, priv->plot_box.y + priv->plot_box.height );
         pango_cairo_show_layout (priv->cr, layout);
    }

    pango_tab_array_free (p_tabs);
    g_object_unref (layout);

    return;
}

/*
 * Draw the chart y scale legend
*/
static void glg_line_graph_draw_y_grid_labels (GlgLineGraph *graph)
{
	GlgLineGraphPrivate *priv;
    gchar       ch_grid_label[GLG_MAX_BUFFER];
    gchar       ch_work[GLG_MAX_BUFFER];
    PangoLayout *layout = NULL;
    gint        y_adj = 0, width = 0, height = 0, v_index = 0;

    g_debug ("===> glg_line_graph_draw_y_grid_labels()");

	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

	priv = graph->priv;
    
    g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "<span font_desc=\"Monospace 8\">%d</span>",
                priv->y_range.i_max_scale);
    layout = pango_cairo_create_layout (priv->cr);

    pango_layout_set_markup (layout, ch_grid_label, -1);
    pango_layout_get_pixel_size (layout, &width, &height);
    y_adj = height / 2;

    g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "<span font_desc=\"Monospace 8\">%d", priv->y_range.i_max_scale);
    for (v_index =
         priv->y_range.i_max_scale - priv->y_range.i_inc_major_scale_by;
         v_index > 0; v_index -= priv->y_range.i_inc_major_scale_by)
    {
        g_strlcpy (ch_work, ch_grid_label, GLG_MAX_BUFFER);
        g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "%s\n%d", ch_work, v_index);
    }
    g_strlcpy (ch_work, ch_grid_label, GLG_MAX_BUFFER);
    g_snprintf (ch_grid_label, GLG_MAX_BUFFER, "%s</span>", ch_work);

    pango_layout_set_spacing (layout, ((priv->y_range.i_major_inc - height) * PANGO_SCALE));
    pango_layout_set_alignment (layout, PANGO_ALIGN_RIGHT);
    pango_layout_set_markup (layout, ch_grid_label, -1);

    pango_cairo_update_layout (priv->cr, layout);

    cairo_set_source_rgba (priv->cr, (gdouble)priv->scale_color.red,
    								 (gdouble)priv->scale_color.green,
    								 (gdouble)priv->scale_color.blue, 0.6);
		 
    cairo_move_to (priv->cr, priv->plot_box.x - (width * 1.4), priv->plot_box.y - y_adj );
    pango_cairo_show_layout (priv->cr, layout);

    g_object_unref (layout);

    return;
}

/*
 * Draws the tooltip legend message at top or bottom of chart
 * returns the width of the text area, or -1 on error
 * requires priv->b_tooltip_active to be TRUE, (toggled by mouse)
*/
static gint glg_line_graph_draw_tooltip (GlgLineGraph *graph)
{
	GlgLineGraphPrivate *priv;
    PangoLayout *layout = NULL;
    gint        x_pos = 0, y_pos = 0, width = 0, height = 0;
    gint        v_index = 0, x_adj = 0;
    PGLG_SERIES  psd = NULL;
    GList      *data_sets = NULL;
    gboolean    b_found = FALSE;
    gdouble     mx=0.0, my=0.0, d_y_match = 0.0, d_value_y = 0.0;
  	gint        x = 0, y = 0;


    g_debug ("===> glg_line_graph_draw_tooltip()");
	
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);

	priv = graph->priv;

    if (!priv->b_tooltip_active)
    {
        return -1;
    }
    if (priv->i_points_available < 1)
    {
        return -1;
    }
	if (priv->cr == NULL) {  /* available during expose only */
		return 1;
	}

    /*
     * Create tooltip if needed */     
    x_adj = (priv->plot_box.width / priv->x_range.i_max_scale);    

	/*
	 * get current mouse pointer, 
	 * and as a side effect allow another notify message */    
  	gdk_window_get_device_position (priv->window, priv->device_pointer, &x, &y, &priv->mouse_state);

    /* 
     * see if mouse ptr is in plot_box x-range point */
     mx = priv->mouse_pos.x = x;
     my = priv->mouse_pos.y = y;
    cairo_device_to_user (priv->cr, &mx, &my);
     priv->mouse_pos.x = (gint) mx;
     priv->mouse_pos.y = (gint) my;     
	if ((priv->mouse_pos.x >= priv->plot_box.x) && (priv->mouse_pos.x <= priv->plot_box.x+priv->plot_box.width) &&
	    (priv->mouse_pos.y >= priv->plot_box.y) && (priv->mouse_pos.y <= priv->plot_box.y+priv->plot_box.height)) {
	    v_index = 0; /* dummy */
	}  else {
	    return -1;
	}   

    for (v_index = 0; v_index <= priv->x_range.i_max_scale; v_index++)
    {
        x_pos = priv->plot_box.x + (v_index * x_adj);
        if ((priv->mouse_pos.x > (x_pos - (x_adj / 3))) &&
            (priv->mouse_pos.x < (x_pos + (x_adj / 3))))
        {
            if (v_index < priv->i_points_available)
            {
                b_found = TRUE;
                d_y_match = priv->mouse_pos.y;
                break;  /* maybe send signal of mouse click here */
            }
        }
    }

    /* 
     * All we needed was x, so now post a tooltip */
    if (b_found)
    {
        gchar       ch_buffer[GLG_MAX_BUFFER];
        gchar       ch_work[GLG_MAX_BUFFER];
        gchar       ch_time_r[GLG_MAX_STRING];
        gchar      *pch_time = NULL;
        time_t      point_time;

        point_time = (time_t) g_list_nth_data (priv->lg_series_time, v_index);

#if _WIN32
        ctime_s(ch_time_r, GLG_MAX_STRING, &point_time);
        pch_time = ch_time_r;
#else
        pch_time = ctime_r (&point_time, ch_time_r);
#endif

        g_strdelimit (pch_time, "\n", ' ');

        g_snprintf (ch_buffer, sizeof (ch_buffer),
                    "<small>{ <u>sample #%d @ %s</u>}\n", v_index, pch_time);
                    
        data_sets = g_list_first (priv->lg_series);
        while (data_sets)
        {
            psd = data_sets->data;
            if (psd != NULL)
            {                   /* found */
                g_snprintf (ch_work, sizeof (ch_work), "%s", ch_buffer);
                g_snprintf (ch_buffer, sizeof (ch_buffer),
                            "%s{%3.2lf <span foreground=\"%s\">%s</span>}",
                            ch_work,
                            psd->lg_point_dvalue[v_index],
                            psd->ch_legend_color, psd->ch_legend_text);
                
                /* compute y point pixel value */                
			    d_value_y = priv->plot_box.y + 
                               (priv->plot_box.height - 
                               (psd->lg_point_dvalue[v_index] * 
                               (gdouble) ((gdouble) priv->plot_box.height / 
                               (gdouble) priv->y_range.i_max_scale)));    
                    
                 /*
                  * if y pixel pos is found emit found signal to subscribers
                  * -- may send more than one based on range matching +-2p 
                  */                            
                if ((d_value_y >= d_y_match - 2) && 
                	(d_value_y <= d_y_match + 2) &&
                	(psd->lg_point_dvalue[v_index] > 0.0 )) {
                	 
                	g_signal_emit (graph, glg_line_graph_signals[POINT_SELECTED_SIGNAL], 0,
                				   	    (gdouble)v_index, 
                				   		(gdouble)psd->lg_point_dvalue[v_index], 
                				   		(gdouble)d_value_y, 
                				   		(gdouble)d_y_match);                	
                }            
                d_value_y = 0.0;
            }
            data_sets = g_list_next (data_sets);
        }

        g_snprintf (ch_work, sizeof (ch_work), "%s", ch_buffer);
        g_snprintf (ch_buffer, sizeof (ch_buffer), "%s</small>", ch_work);
        glg_line_graph_chart_set_text (graph, GLG_TOOLTIP, ch_buffer);
    }

    if ( (!b_found) )
    {
        return -1;
    }

    layout = pango_cairo_create_layout (priv->cr);
    pango_layout_set_markup (layout, priv->ch_tooltip_text, -1);
    pango_layout_set_alignment (layout, PANGO_ALIGN_CENTER);
    pango_layout_get_pixel_size (layout, &width, &height);

    x_pos = priv->tooltip_box.x + ((priv->tooltip_box.width - width) / 2);
    y_pos = priv->tooltip_box.y + ((priv->tooltip_box.height - height) / 2);
    /* box_gc */
    cairo_set_source_rgb (priv->cr, (gdouble)priv->window_color.red,
    								 (gdouble)priv->window_color.green,
    								 (gdouble)priv->window_color.blue);
    cairo_rectangle (priv->cr, priv->tooltip_box.x, priv->tooltip_box.y, 
    						   priv->tooltip_box.width, priv->tooltip_box.height);
    cairo_fill (priv->cr);
    cairo_set_source_rgb (priv->cr, (gdouble)priv->scale_color.red,
    								 (gdouble)priv->scale_color.green,
    								 (gdouble)priv->scale_color.blue);
    cairo_stroke (priv->cr);

    cairo_set_source_rgba (priv->cr, (gdouble)priv->scale_color.red,
    								 (gdouble)priv->scale_color.green,
    								 (gdouble)priv->scale_color.blue, 1.0);
    cairo_move_to (priv->cr, x_pos, y_pos);
    pango_cairo_show_layout (priv->cr, layout);

    g_object_unref (layout);


    return width;
}

/*
 * Draws one data series points to chart
 * returns number of points processed
*/
static gint glg_line_graph_data_series_draw (GlgLineGraph *graph, PGLG_SERIES psd)
{
	GlgLineGraphPrivate *priv;
    gint        v_index = 0;
    GdkPoint   *point_pos = NULL;

    g_debug ("===> glg_line_graph_data_series_draw(entered)");
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), -1);
    g_return_val_if_fail (psd != NULL, -1);
    g_return_val_if_fail (psd->point_pos != NULL, -1);

	priv = graph->priv;

	if (priv->cr == NULL) {  /* available during expose only */
		return 1;
	}

    cairo_set_source_rgb (priv->cr, (gdouble)psd->legend_color.red,
                                     (gdouble)psd->legend_color.green,
                                     (gdouble)psd->legend_color.blue);
    cairo_set_line_width (priv->cr, (gdouble)priv->series_line_width);
    cairo_set_line_cap (priv->cr, CAIRO_LINE_CAP_ROUND);

    point_pos = psd->point_pos;

/* trap first and only point */
    if (psd->i_point_count == 0)
    {
        return 0;
    }
    if (psd->i_point_count == 1)
    {
        point_pos[0].x = priv->plot_box.x;
        point_pos[0].y =
            priv->plot_box.y + (priv->plot_box.height -
            (psd->lg_point_dvalue[0] *
              (gdouble) ((gdouble) priv->plot_box.height /
                         (gdouble) priv->y_range.i_max_scale)));

		cairo_move_to (priv->cr, point_pos[0].x, point_pos[0].y);
		cairo_arc (priv->cr, point_pos[0].x, point_pos[0].y, 3., 0., 2 * M_PI);
		
		cairo_fill(priv->cr);
        return 1;
    }

    for (v_index = 0; v_index < psd->i_point_count; v_index++)
    {
        point_pos[v_index].x = priv->plot_box.x + (v_index * (priv->plot_box.width / priv->x_range.i_max_scale));
        point_pos[v_index].y = priv->plot_box.y + 
                               (priv->plot_box.height - 
                               (psd->lg_point_dvalue[v_index] * 
                               (gdouble) ((gdouble) priv->plot_box.height / (gdouble) priv->y_range.i_max_scale)));
                                       
		if (v_index == 0) {
		    cairo_move_to (priv->cr, point_pos[v_index].x, point_pos[v_index].y);
		} else {
			cairo_line_to (priv->cr, point_pos[v_index].x, point_pos[v_index].y);
		}
    }
    cairo_stroke (priv->cr);

	cairo_set_line_width (priv->cr, 2.0);
    for (v_index = 0; v_index < psd->i_point_count ; v_index++)
    {
		 cairo_move_to (priv->cr, point_pos[v_index].x , point_pos[v_index].y );
		 cairo_arc (priv->cr, point_pos[v_index].x , point_pos[v_index].y , 3.0, 0., 2 * M_PI);
    }
	 cairo_fill(priv->cr);

    return v_index;
}

/*
 * Draws all data series points to chart
 * returns number of series processed
*/
static gint glg_line_graph_data_series_draw_all (GlgLineGraph *graph, gboolean redraw_control)
{
	GlgLineGraphPrivate *priv;
    PGLG_SERIES  psd = NULL;
    GList      *data_sets = NULL;
    gint        v_index = 0, points = 0;
    gint64 start_time = glg_duration_us(NULL, NULL);
    gchar buff[64];

    g_debug ("===> glg_line_graph_data_series_draw_all(entered)");
	
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);

	priv = graph->priv;
    

    data_sets = g_list_first (priv->lg_series);
    while (data_sets)
    {
        psd = data_sets->data;
        if (psd != NULL)
        {                       /* found */
        points = glg_line_graph_data_series_draw (graph, psd);
        	g_snprintf(buff, sizeof(buff), "glg_line_graph_data_series_draw#[%d]Series", v_index);
        	glg_duration_us(&start_time, buff);

            v_index++;
        }
        data_sets = g_list_next (data_sets);
    }

    g_debug ("glg_line_graph_data_series_draw_all(exited): #series=%d, #points=%d", v_index, points);

    return v_index;
}

/**
 * glg_line_graph_data_series_add_value:
 * @graph: pointer to a #GlgLineGraph widget
 * @i_series_number: The data series to add value to.
 * @y_value: The value to add.  Current range is 0.0 to y-max-scale
 * 
 * Add a single y value to the requested data series.
 * auto indexes the value if x-scale max is reached (appends to the end)
 * The X value is implied to be the current count of Y-values added.
 *
 * Returns: gboolean  TRUE if value was added, FALSE if add failed.
 */
extern gboolean glg_line_graph_data_series_add_value (GlgLineGraph *graph, gint i_series_number, gdouble y_value)
{
	GlgLineGraphPrivate *priv;
    PGLG_SERIES  psd = NULL;
    GList      *data_sets = NULL;
    gint        v_index = 0;
    gboolean    b_found = FALSE;

    g_debug ("===> glg_line_graph_data_series_add_value()");

	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);
	g_return_val_if_fail (gtk_widget_get_realized (GTK_WIDGET(graph)), FALSE);

	priv = graph->priv;
    
    data_sets = g_list_first (priv->lg_series);  /* Find specified series */
    while (data_sets)
    {
        psd = data_sets->data;
        if (psd->i_series_id == i_series_number)
        {                       /* found */
            b_found = TRUE;
            break;
        }
        data_sets = g_list_next (data_sets);
    }

    if (!b_found)
    {
        g_message ("glg_line_graph_data_series_add_value(%d): Invalid data series number",
                   i_series_number);
        return FALSE;
    }

    if (y_value > priv->y_range.i_max_scale)
    {
        y_value = (gdouble) priv->y_range.i_max_scale;
    }

    if (psd->i_point_count == psd->i_max_points + 1)
    {
        for (v_index = 0; v_index < psd->i_max_points; v_index++)
        {
            psd->lg_point_dvalue[v_index] = psd->lg_point_dvalue[v_index + 1];
        }
        psd->lg_point_dvalue[psd->i_max_points] = y_value;
    }
    else
    {
        psd->lg_point_dvalue[psd->i_point_count++] = y_value;
    }

    psd->d_max_value = MAX (y_value, psd->d_max_value);
    psd->d_min_value = MIN (y_value, psd->d_min_value);

    priv->i_points_available = MAX (priv->i_points_available, psd->i_point_count);

    /* record current time with data points */
    if (psd->i_series_id == priv->i_num_series - 1)
    {
        GList *gl_remove = NULL;

        if (g_list_length (priv->lg_series_time) == (guint)psd->i_max_points +1 )
        {
            gl_remove = g_list_first (priv->lg_series_time);
            	priv->lg_series_time = g_list_remove (priv->lg_series_time, gl_remove->data);
        }
        priv->lg_series_time =
            g_list_append (priv->lg_series_time, GINT_TO_POINTER ((time_t) time (NULL)));
            /* TODO: Leaking Memory - NO time_t is a gint64
                     time always 1970 */
    }

    g_debug ("  ==>DataSeriesAddValue: series=%d, value=%3.1lf, index=%d, count=%d, max_pts=%d",
             i_series_number, y_value, v_index, psd->i_point_count, psd->i_max_points);

    return TRUE;
}

/*
 * A shutdown routine
 * destroys all the data series and any assocaited dynamic data
*/
static gint glg_line_graph_data_series_remove_all (GlgLineGraph *graph)
{
	GlgLineGraphPrivate *priv;
    PGLG_SERIES  psd = NULL;
    GList      *data_sets = NULL;
    gint        i_count = 0;

    g_debug ("===> glg_line_graph_data_series_remove_all()");
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);

	priv = graph->priv;
    
    data_sets = g_list_first (priv->lg_series);
    while (data_sets)
    {
        psd = data_sets->data;
        g_free (psd->lg_point_dvalue);
        g_free (psd->point_pos);
        data_sets = g_list_next (data_sets);
        i_count++;
    }
    g_list_free_full (priv->lg_series, g_free);
    g_list_free (priv->lg_series_time);
    priv->lg_series = NULL;
    priv->lg_series_time = NULL;
    priv->i_num_series = 0;
    priv->i_points_available = 0;    

    g_debug ("  ==>DataSeriesRemoveAll: number removed=%d", i_count);

    return TRUE;
}

/**
 * glg_line_graph_data_series_add:
 * @graph: pointer to a #GlglineGraph widget
 * @pch_legend_text: The name of the data series
 * @pch_color_text: A string containing the line color to be used.
 *
 * Allocates space for another data series of y-values and returns the 
 * series number of this dataset added which you must keep track of
 * to add values.
 *
 * Returns: gint  The series number of this dataset added ( range 0 thru n )
 */
extern gint glg_line_graph_data_series_add (GlgLineGraph *graph, const gchar *pch_legend_text, const gchar *pch_color_text)
{
	GlgLineGraphPrivate *priv;
    PGLG_SERIES  psd = NULL;

    g_debug ("===> glg_line_graph_data_series_add()");
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), FALSE);
    g_return_val_if_fail (pch_legend_text != NULL, -1);
    g_return_val_if_fail (pch_color_text != NULL, -1);

	priv = graph->priv;

    psd = (PGLG_SERIES) g_new0 (GLG_SERIES, 1);
    g_return_val_if_fail (psd != NULL, -1);

    psd->lg_point_dvalue = (gdouble *) g_new0 (gdouble, (priv->x_range.i_max_scale + 4));
    g_return_val_if_fail (psd->lg_point_dvalue != NULL, -1);

    psd->point_pos = g_new0 (GdkPoint, (priv->x_range.i_max_scale + 4));
    g_return_val_if_fail (psd->point_pos != NULL, -1);

    g_snprintf (psd->ch_legend_text, sizeof (psd->ch_legend_text), "%s", pch_legend_text);
    
    /*
     * we position x to ticks onlys, 
     * so force chart to scroll at maximum ticks vs value
     * psd->i_max_points = MIN (priv->x_range.i_max_scale, priv->x_range.i_num_minor); */
    psd->i_max_points = priv->x_range.i_max_scale;
        
    gdk_rgba_parse (&psd->legend_color, pch_color_text);
    g_snprintf (psd->ch_legend_color, sizeof (psd->ch_legend_color), "%s", pch_color_text);
    psd->cb_id = GLG_SERIES_ID;

    priv->lg_series = g_list_append (priv->lg_series, psd);
    psd->i_series_id = priv->i_num_series++;

    g_debug ("  ==>DataSeriesAdd: series=%d, max_pts=%d", psd->i_series_id, psd->i_max_points);

    return psd->i_series_id;
}

/*
 * Toggle the legend function on off
 * "button-press-event"
 * secret decoder function: button 3 enables, button 2, for debug messages toggle
*/
static gboolean glg_line_graph_button_press_event (GtkWidget * widget, GdkEventButton * ev)
{
    GlgLineGraphPrivate *priv;

	priv = GLG_LINE_GRAPH(widget)->priv;

	if ( !(((ev->x >= priv->plot_box.x) &&      // filter out moves not inside plot box
		    (ev->y >= priv->plot_box.y)) &&
		   ((ev->x <= priv->plot_box.x + priv->plot_box.width) &&      // filter out moves not inside plot box
		    (ev->y <= priv->plot_box.y + priv->plot_box.height)))
	   ) {
		return TRUE;
	}

    g_debug ("===> glg_line_graph_button_press_event_cb()");
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(widget), TRUE);
    
    if ((ev->type & GDK_BUTTON_PRESS) && (ev->button == 1))
    {
        priv->b_tooltip_active = priv->b_tooltip_active ? FALSE : TRUE;    	
        gdk_window_get_device_position (ev->window, priv->device_pointer, &priv->mouse_pos.x, &priv->mouse_pos.y, &priv->mouse_state);
        glg_line_graph_redraw (GLG_LINE_GRAPH(widget));        /* point select action */
        return TRUE;
    }
    if ((ev->type & GDK_BUTTON_PRESS) && (ev->button == 3))
    {
        priv->b_mouse_onoff = priv->b_mouse_onoff ? FALSE : TRUE;
        glg_line_graph_redraw (GLG_LINE_GRAPH(widget));        /* point select action */
        return TRUE;        
    }

    return FALSE;
}

/*
 * Track the mouse pointer position
 * "motion-notify-event"
*/
static gboolean glg_line_graph_motion_notify_event (GtkWidget * widget, GdkEventMotion * ev)
{
  GlgLineGraphPrivate *priv;
  GdkModifierType state;
  gint        x = 0, y = 0;

  priv = GLG_LINE_GRAPH(widget)->priv;

	if ( !(((ev->x >= priv->plot_box.x) &&      // filter out moves not inside plot box
		    (ev->y >= priv->plot_box.y)) &&
		   ((ev->x <= priv->plot_box.x + priv->plot_box.width) &&      // filter out moves not inside plot box
		    (ev->y <= priv->plot_box.y + priv->plot_box.height)))
	   ) {
		return TRUE;
	}

    g_debug ("===> glg_line_graph_motion_notify_event_cb()");

    g_return_val_if_fail ( GLG_IS_LINE_GRAPH(widget), TRUE);

    if (ev->is_hint)
    {
        gdk_window_get_device_position (ev->window, priv->device_pointer, &x, &y, &state);
    }
    else
    {
        x = ev->x;
        y = ev->y;
        state = ev->state;
    }

    /* save device coordinates */
    priv->mouse_pos.x = x;
    priv->mouse_pos.y = y;
    priv->mouse_state = state;

    if (( priv->lgflags & GLG_TOOLTIP) && priv->b_tooltip_active ) {     
          glg_line_graph_redraw (GLG_LINE_GRAPH(widget)); 
    }
    
    return TRUE;
}

static void glg_line_graph_destroy (GtkWidget *object)
{
  GlgLineGraphPrivate *priv = NULL;
  GtkWidget       *widget = NULL;

  g_debug ("===> glg_line_graph_destroy(enter)");

  g_return_if_fail (object != NULL);
  
  widget = GTK_WIDGET( object );

  g_return_if_fail ( GLG_IS_LINE_GRAPH(widget));

  priv = GLG_LINE_GRAPH(widget)->priv;
  g_return_if_fail ( priv != NULL );
  
  if ( priv->x_label_text )  /* avoid multiple destroys */
  {
      glg_line_graph_data_series_remove_all ( GLG_LINE_GRAPH( widget ) );

      g_free(priv->x_label_text);
      g_free(priv->y_label_text);
      g_free(priv->page_title_text);
      priv->x_label_text = NULL;
      priv->y_label_text = NULL;
      priv->page_title_text = NULL;

      cairo_surface_destroy(priv->surface);

      if (GTK_WIDGET_CLASS (glg_line_graph_parent_class)->destroy != NULL)
      {
         (*GTK_WIDGET_CLASS (glg_line_graph_parent_class)->destroy) (object);
      }
  }

  g_debug ("glg_line_graph_destroy(exited)");

  return;
}

/**
 * glg_line_graph_chart_set_elements:
 * @graph: pointer to a #GlglineGraph widget
 * @element: An or'ed list of #GLGElementID indicating what graph elements should be drawn
 * to the screen. 
 *
 * Controls whether the grids, labels, tooltip, and titles will appear on the chart.
 * All graphs are created with the following defaults; 
 * %GLG_TOOLTIP | 
 * %GLG_GRID_LABELS_X | %GLG_GRID_LABELS_Y | %GLG_TITLE_T | %GLG_TITLE_X | %GLG_TITLE_Y | 
 * %GLG_GRID_LINES | %GLG_GRID_MINOR_X | %GLG_GRID_MAJOR_X | %GLG_GRID_MINOR_Y | %GLG_GRID_MAJOR_Y 
 */
extern void glg_line_graph_chart_set_elements ( GlgLineGraph *graph, GLGElementID element)
{
	
    g_debug ("===> glg_line_graph_chart_set_elements(entered)");
	g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

  	g_return_if_fail ( graph->priv != NULL );

    graph->priv->lgflags |= element;

    g_debug ("===> glg_line_graph_chart_set_elements(exited)");
}

/**
 * glg_line_graph_chart_get_elements:
 * @graph: pointer to a GlglineGraph widget
 * 
 * Retrieves the current draw setting for the graph.  AND it with the desired value to test
 * if what your after is present.
 *
 * Returns: gint containing all current draw flags as defined in #GLGElementID
 */
extern GLGElementID glg_line_graph_chart_get_elements ( GlgLineGraph *graph)
{
    
    g_debug ("===> glg_line_graph_chart_get_elements(entered)");
	g_return_val_if_fail ( GLG_IS_LINE_GRAPH(graph), 0);

  	g_return_val_if_fail ( graph->priv != NULL, 0 );

    g_debug ("===> glg_line_graph_chart_get_elements(exited)");
    return graph->priv->lgflags;
}


static void glg_line_graph_set_property (GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec)
{
  GlgLineGraphPrivate *priv = NULL;
  GlgLineGraph *graph = NULL;
  gint  i_value = 0;


  g_debug ("===> glg_line_graph_set_property(entered)");
  g_return_if_fail ( object != NULL);    

  graph = GLG_LINE_GRAPH (object);
  g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

  priv = graph->priv;
  g_return_if_fail ( priv != NULL);
  
  switch (prop_id)
    {
    case PROP_GRAPH_TITLE:
      glg_line_graph_chart_set_text (graph, GLG_TITLE_T, g_value_get_string(value) );
      break;      
    case PROP_GRAPH_TITLE_X:
      glg_line_graph_chart_set_text (graph, GLG_TITLE_X, g_value_get_string(value) );
      break;      
    case PROP_GRAPH_TITLE_Y:
      glg_line_graph_chart_set_text (graph, GLG_TITLE_Y, g_value_get_string(value) );
      break;      
    case PROP_GRAPH_LINE_WIDTH:
      priv->series_line_width = g_value_get_int ( value );
      break;
    case PROP_GRAPH_ELEMENTS:
      priv->lgflags |= g_value_get_int ( value );    
      break;
	case PROP_GRAPH_TITLE_COLOR:
      glg_line_graph_chart_set_color (graph, GLG_TITLE,  g_value_get_string(value));
	  break;      
	case PROP_GRAPH_SCALE_COLOR:
      glg_line_graph_chart_set_color (graph, GLG_SCALE,  g_value_get_string(value));
	  break;      
	case PROP_GRAPH_CHART_COLOR:
      glg_line_graph_chart_set_color (graph, GLG_CHART,  g_value_get_string(value));
	  break;      
	case PROP_GRAPH_WINDOW_COLOR:
      glg_line_graph_chart_set_color (graph, GLG_WINDOW, g_value_get_string(value));	
	  break;    
	case PROP_TICK_MINOR_X:
	  i_value = g_value_get_int ( value );
	  priv->x_range.i_inc_minor_scale_by = i_value;
	  priv->x_range.i_num_minor = priv->x_range.i_max_scale / i_value;
	  break;	    
	case PROP_TICK_MAJOR_X:
	  i_value = g_value_get_int ( value );
      priv->x_range.i_inc_major_scale_by = i_value;	  	
      priv->x_range.i_num_major = priv->x_range.i_max_scale / i_value;
	  break;	    
	case PROP_SCALE_MINOR_X:
	  i_value = g_value_get_int ( value );
	  priv->x_range.i_min_scale = i_value;	
	  break;	    
	case PROP_SCALE_MAJOR_X:
	  i_value = g_value_get_int ( value );
	  if (priv->x_range.i_max_scale) {
		  g_message ("Set Properties Failed: Cannot set ranges more than once, range already set!");
		  break;
	  }	  
	  priv->x_range.i_max_scale = i_value;	
      priv->x_range.i_num_minor = i_value / priv->x_range.i_inc_minor_scale_by;
      priv->x_range.i_num_major = i_value / priv->x_range.i_inc_major_scale_by; 
	  break;	    
	case PROP_TICK_MINOR_Y:
	  i_value = g_value_get_int ( value );
	  priv->y_range.i_inc_minor_scale_by = i_value;
	  priv->y_range.i_num_minor = priv->y_range.i_max_scale / i_value;
	  break;	    
	case PROP_TICK_MAJOR_Y:
	  i_value = g_value_get_int ( value );
      priv->y_range.i_inc_major_scale_by = i_value;	  	
      priv->y_range.i_num_major = priv->y_range.i_max_scale / i_value;
	  break;	    
	case PROP_SCALE_MINOR_Y:
	  i_value = g_value_get_int ( value );
	  priv->y_range.i_min_scale = i_value;	
	  break;	    
	case PROP_SCALE_MAJOR_Y:
	  i_value = g_value_get_int ( value );
	  priv->y_range.i_max_scale = i_value;	
      priv->y_range.i_num_minor = i_value / priv->y_range.i_inc_minor_scale_by;
      priv->y_range.i_num_major = i_value / priv->y_range.i_inc_major_scale_by; 
	  break;	    

    default:
      G_OBJECT_WARN_INVALID_PROPERTY_ID (object, prop_id, pspec);
      break;
    }

    g_debug ("===> glg_line_graph_set_property(exited)");
    
    return;
}

static void glg_line_graph_get_property (GObject *object, guint prop_id, GValue *value, GParamSpec *pspec)
{
  GlgLineGraphPrivate *priv = NULL;
  GlgLineGraph *graph = NULL;

  g_debug ("===> glg_line_graph_get_property(entered)");
  
  g_return_if_fail ( object != NULL);    

  graph = GLG_LINE_GRAPH (object);
  g_return_if_fail ( GLG_IS_LINE_GRAPH(graph));

  priv = graph->priv;
  g_return_if_fail ( priv != NULL);
  
  switch (prop_id)
    {
    case PROP_GRAPH_ELEMENTS:
          g_value_set_int (value, priv->lgflags);
      break;
    default:
      G_OBJECT_WARN_INVALID_PROPERTY_ID (object, prop_id, pspec);
      break;
    }

    g_debug ("===> glg_line_graph_get_property(exited)");
    
    return;
}

/*
 * GOBJECT Marshalling routines
 * - required by class init and point-selected-signal 
*/

#ifdef G_ENABLE_DEBUG
	#define g_marshal_value_peek_double(v)   g_value_get_double (v)
#else /* !G_ENABLE_DEBUG */
	#define g_marshal_value_peek_double(v)   (v)->data[0].v_double
#endif /* !G_ENABLE_DEBUG */

static void _glg_cairo_marshal_VOID__DOUBLE_DOUBLE_DOUBLE_DOUBLE (GClosure     *closure,
                                                      GValue       *return_value,
                                                      guint         n_param_values,
                                                      const GValue *param_values,
                                                      gpointer      invocation_hint,
                                                      gpointer      marshal_data)
{
  typedef void (*GMarshalFunc_VOID__DOUBLE_DOUBLE_DOUBLE_DOUBLE) (gpointer     data1,
                                                                  gdouble      arg_1,
                                                                  gdouble      arg_2,
                                                                  gdouble      arg_3,
                                                                  gdouble      arg_4,
                                                                  gpointer     data2);
  register GMarshalFunc_VOID__DOUBLE_DOUBLE_DOUBLE_DOUBLE callback;
  register GCClosure *cc = (GCClosure*) closure;
  register gpointer data1, data2;

  g_return_if_fail (n_param_values == 5);

  if (G_CCLOSURE_SWAP_DATA (closure))
    {
      data1 = closure->data;
      data2 = g_value_peek_pointer (param_values + 0);
    }
  else
    {
      data1 = g_value_peek_pointer (param_values + 0);
      data2 = closure->data;
    }
  callback = (GMarshalFunc_VOID__DOUBLE_DOUBLE_DOUBLE_DOUBLE) (marshal_data ? marshal_data : cc->callback);

  callback (data1,
            g_marshal_value_peek_double (param_values + 1),
            g_marshal_value_peek_double (param_values + 2),
            g_marshal_value_peek_double (param_values + 3),
            g_marshal_value_peek_double (param_values + 4),
            data2);
}
