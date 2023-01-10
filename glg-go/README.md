#GlgLineGraph

This is the original README file of GlgLineGraph library.

Visit [here](https://github.com/skoona/glinegraph-cairo) to fetch the origin repo released under LGPLv2.0.

Thanks [skoona](https://github.com/skoona) for his great work.

### (a.k.a cairo version of glinegraph )  July 2007/2016

![GLineGraph Widget](https://github.com/skoona/glinegraph-cairo/raw/master/images/glg_cairo3.png) 

 A Gtk3/GLib2/cario GUI application which demonstrates the use of GTK+ for producing xy line graphs.  This widget once created allows you to add one or more data series, then add values to those series for ploting.  The X point is assumed based on arrival order.  However, the Y value or position is based one the current scale and the y value itself.  If the charts x scale maximum is 40, or 40 points, the 41+ value is appended to the 40th position after all points are shifted left and out through pos 0.
 
 Packaged as a gtk widget for ease of use.	

 A lgcairo.c example program is included to demonstrate the possible use of the codeset.

There is also [GtkDoc API documentation](https://skoona.github.io/glinegraph-cairo/docs/reference/html/index.html) in ./docs directory.

FEATURES: 

    * GlgLineGraph API Reference Manual in ./docs directory.
    * Unlimited data series support.
    * Accurate scaling across a wide range of X & Y scales.
      - Using values ranges above or below 1.
    * Rolling data points, if number of x points exceed x-scale. (left shift)
    * Ability to change chart background color, window backgrounds colors, etc.
    * Popup Tooltip, via mouse-button-1 click to enable/toggle.
      - Tooltip overlays top graph title, when present.
    * Data points are time stamped with current time when added.
	* Some key debug messages to console: $ export G_MESSAGES_DEBUG=all
	* Auto Size to current window size; i.e. no-scrolling.
	* point-selected signal tied to tooltip, to display y value under mouse.

REQUIREMENTS:

	Gtk3/Glib2 runtime support.

	* the following packages may need to be installed to build program and
	  for runtime of binaries on some platforms.

	glinegraph - { this package }
	gtk-devel  - { GTK+ version 3.10 is minimum package level required}
	glib-devel - { version 2.40, packaged with GTK+ } 
						 		

DISTRIBUTION METHOD:

	Source tar	glinegraph-{version}.tar.bz2
	GPL2 copyright

INSTALL INFO:

	Configure Source with 
			'# ./autogen.sh --enable-gtk-doc '
			'# make clean '
			'# make '
			'# make install'

			-- or --
			
			'# ./autogen.sh --disable-gtk-doc '
			'# make clean '
			'# make '
			'# make install'


INSTRUCTIONS: 

  glinegraph -- GTK+ Cairo Programing Example Application
    
    1. Compile program
	2. Execute sample program
		- regular cmd: "# lgcairo" 


KNOWN BUGS:

    None.
	5/23/2016 Thanks to LinuxQuestions member: norobro for helping debug the performance issue I was having.



BEGIN DEBUG LOG: 

### To see console log -- $ export G_MESSAGES_DEBUG=all

		[jscott OSX El-Capitan GLG-Cairo]$ pkg-config --modversion gtk+-3.0
		#==> 3.16.6
		
		[jscott OSX El-Capitan glinegraph-cairo]$ src/lgcairo 

		** (lgcairo:68503): DEBUG: ===> glg_line_graph_new(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_class_init(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_class_init(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_init(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_init(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_color(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_text(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_text(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_text(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_text(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_text(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_chart_set_text(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_set_property(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_new(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAdd: series=0, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAdd: series=1, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAdd: series=2, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAdd: series=3, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAdd: series=4, max_pts=40
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_size_allocate(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_size_allocate(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_realize(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_send_configure(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_configure_event(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_compute_layout(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_compute_layout(new width=570, height=270)
		** (lgcairo:68503): DEBUG: Alloc:factors:raw:pango_layout_get_pixel_size(width=10, height=12)
		** (lgcairo:68503): DEBUG: Alloc:factors:adj:pango_layout_get_pixel_size(width=10, height=20)
		** (lgcairo:68503): DEBUG: Alloc:Max.Avail: plot_box.width=495, plot_box.height=175
		** (lgcairo:68503): DEBUG: Alloc:Chart:Incs:    x_minor=12, x_major=120, y_minor=3, y_major=15, plot_box.x=77, plot_box.y=60, plot_box.width=480, plot_box.height=165
		** (lgcairo:68503): DEBUG: Alloc:Chart:Nums:    x_num_minor=40, x_num_major=4, y_num_minor=55, y_num_major=11
		** (lgcairo:68503): DEBUG: Alloc:Chart:Plot:    x=77, y=60, width=480, height=165
		** (lgcairo:68503): DEBUG: Alloc:Chart:Title:   x=77, y=5, width=480, height=40
		** (lgcairo:68503): DEBUG: Alloc:Chart:yLabel:  x=5, y=225, width=30, height=150
		** (lgcairo:68503): DEBUG: Alloc:Chart:xLabel:  x=77, y=240, width=480, height=25
		** (lgcairo:68503): DEBUG: Alloc:Chart:Tooltip: x=77, y=5, width=480, height=45
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_compute_layout(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_graph(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#PlotArea() duration=0.281 ms.
		** (lgcairo:68503): DEBUG: Chart.Surface: pg.Width=570, pg.Height=270, Plot Area x=77 y=60 width=480, height=165
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_horizontal()
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Page cx=570, cy=270
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Orig: x=77, y=5, cx=480, cy=40
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Calc x_pos=203, y_pos=8,  cx=228, cy=36
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Top-Title() duration=8.385 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_horizontal()
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Page cx=570, cy=270
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Orig: x=77, y=240, cx=480, cy=25
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Calc x_pos=171, y_pos=247,  cx=291, cy=16
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#X-Title() duration=2.081 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_vertical()
		** (lgcairo:68503): DEBUG: Vert:TextBox: y_pos=219,  x=5, y=225, cx=168, cy=30
		** (lgcairo:68503): DEBUG: Vert.TextBox: y_pos=219,  x=5, y=225, cx=168, cy=30
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Y-Title() duration=3.719 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_grid_lines()
		** (lgcairo:68503): DEBUG: Draw.Y-GridLines: count_major=10, count_minor=54, y_minor_inc=3, y_major_inc=15
		** (lgcairo:68503): DEBUG: Draw.X-GridLines: count_major=3, count_minor=39, x_minor_inc=12, x_major_inc=120
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#GridLines() duration=0.818 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_x_grid_labels()
		** (lgcairo:68503): DEBUG: Scale:Labels:X small font sizes cx=13, cy=11
		** (lgcairo:68503): DEBUG: Scale:Labels:X plot_box.cx=480, layout.cx=493, layout.cy=11
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#X-Labels() duration=7.588 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_y_grid_labels()
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Y-Labels() duration=0.651 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw_all(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[0]Series() duration=0.012 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[1]Series() duration=0.004 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[2]Series() duration=0.003 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[3]Series() duration=0.003 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[4]Series() duration=0.004 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_data_series_draw_all(exited): #series=5, #points=0
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Series-All() duration=0.044 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_tooltip()
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Tooltip() duration=0.003 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_graph(exited)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#TOTAL-TIME() duration=23.651 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_configure_event(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_send_configure(exited)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_realize(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_master_draw(entered)
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(Allocation ==> width=570, height=270,  Dirty Rect ==> x=0, y=0, width=570, height=270 )
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_master_draw#TOTAL-TIME() duration=2.810 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_master_draw(entered)
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(Allocation ==> width=570, height=270,  Dirty Rect ==> x=0, y=0, width=570, height=270 )
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_master_draw#TOTAL-TIME() duration=2.357 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=0, value=17.4, index=0, count=1, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=1, value=26.1, index=0, count=1, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=2, value=82.7, index=0, count=1, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=3, value=53.3, index=0, count=1, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=4, value=99.6, index=0, count=1, max_pts=40
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_redraw(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_graph(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#PlotArea() duration=0.747 ms.
		** (lgcairo:68503): DEBUG: Chart.Surface: pg.Width=570, pg.Height=270, Plot Area x=77 y=60 width=480, height=165
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_horizontal()
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Page cx=570, cy=270
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Orig: x=77, y=5, cx=480, cy=40
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Calc x_pos=203, y_pos=8,  cx=228, cy=36
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Top-Title() duration=0.448 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_horizontal()
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Page cx=570, cy=270
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Orig: x=77, y=240, cx=480, cy=25
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Calc x_pos=171, y_pos=247,  cx=291, cy=16
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#X-Title() duration=0.298 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_vertical()
		** (lgcairo:68503): DEBUG: Vert:TextBox: y_pos=219,  x=5, y=225, cx=168, cy=30
		** (lgcairo:68503): DEBUG: Vert.TextBox: y_pos=219,  x=5, y=225, cx=168, cy=30
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Y-Title() duration=0.463 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_grid_lines()
		** (lgcairo:68503): DEBUG: Draw.Y-GridLines: count_major=10, count_minor=54, y_minor_inc=3, y_major_inc=15
		** (lgcairo:68503): DEBUG: Draw.X-GridLines: count_major=3, count_minor=39, x_minor_inc=12, x_major_inc=120
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#GridLines() duration=1.428 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_x_grid_labels()
		** (lgcairo:68503): DEBUG: Scale:Labels:X small font sizes cx=13, cy=11
		** (lgcairo:68503): DEBUG: Scale:Labels:X plot_box.cx=480, layout.cx=493, layout.cy=11
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#X-Labels() duration=0.329 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_y_grid_labels()
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Y-Labels() duration=0.420 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw_all(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[0]Series() duration=0.061 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[1]Series() duration=0.028 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[2]Series() duration=0.026 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[3]Series() duration=0.025 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[4]Series() duration=0.024 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_data_series_draw_all(exited): #series=5, #points=1
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Series-All() duration=0.203 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_tooltip()
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Tooltip() duration=0.025 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_graph(exited)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#TOTAL-TIME() duration=4.476 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_redraw(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_master_draw(entered)
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(Allocation ==> width=570, height=270,  Dirty Rect ==> x=0, y=0, width=570, height=270 )
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_master_draw#TOTAL-TIME() duration=3.057 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(exited)
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=0, value=9.8, index=0, count=2, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=1, value=20.4, index=0, count=2, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=2, value=82.7, index=0, count=2, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=3, value=79.1, index=0, count=2, max_pts=40
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_add_value()
		** (lgcairo:68503): DEBUG:   ==>DataSeriesAddValue: series=4, value=99.3, index=0, count=2, max_pts=40
		
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_redraw(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_graph(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#PlotArea() duration=0.858 ms.
		** (lgcairo:68503): DEBUG: Chart.Surface: pg.Width=570, pg.Height=270, Plot Area x=77 y=60 width=480, height=165
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_horizontal()
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Page cx=570, cy=270
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Orig: x=77, y=5, cx=480, cy=40
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Calc x_pos=203, y_pos=8,  cx=228, cy=36
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Top-Title() duration=0.458 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_horizontal()
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Page cx=570, cy=270
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Orig: x=77, y=240, cx=480, cy=25
		** (lgcairo:68503): DEBUG: Horiz.TextBox:Calc x_pos=171, y_pos=247,  cx=291, cy=16
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#X-Title() duration=0.303 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_text_vertical()
		** (lgcairo:68503): DEBUG: Vert:TextBox: y_pos=219,  x=5, y=225, cx=168, cy=30
		** (lgcairo:68503): DEBUG: Vert.TextBox: y_pos=219,  x=5, y=225, cx=168, cy=30
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Y-Title() duration=0.472 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_grid_lines()
		** (lgcairo:68503): DEBUG: Draw.Y-GridLines: count_major=10, count_minor=54, y_minor_inc=3, y_major_inc=15
		** (lgcairo:68503): DEBUG: Draw.X-GridLines: count_major=3, count_minor=39, x_minor_inc=12, x_major_inc=120
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#GridLines() duration=1.446 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_x_grid_labels()
		** (lgcairo:68503): DEBUG: Scale:Labels:X small font sizes cx=13, cy=11
		** (lgcairo:68503): DEBUG: Scale:Labels:X plot_box.cx=480, layout.cx=493, layout.cy=11
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#X-Labels() duration=0.304 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_y_grid_labels()
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Y-Labels() duration=0.420 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw_all(entered)
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[0]Series() duration=0.078 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[1]Series() duration=0.057 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[2]Series() duration=0.054 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[3]Series() duration=0.058 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_data_series_draw(entered)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_data_series_draw#[4]Series() duration=0.055 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_data_series_draw_all(exited): #series=5, #points=2
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Series-All() duration=0.344 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_tooltip()
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#Tooltip() duration=0.004 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_draw_graph(exited)
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_draw_graph#TOTAL-TIME() duration=4.719 ms.
		** (lgcairo:68503): DEBUG: ===> glg_line_graph_redraw(exited)

		** (lgcairo:68503): DEBUG: ===> glg_line_graph_master_draw(entered)
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(Allocation ==> width=570, height=270,  Dirty Rect ==> x=0, y=0, width=570, height=270 )
		** (lgcairo:68503): DEBUG: DURATION: glg_line_graph_master_draw#TOTAL-TIME() duration=3.068 ms.
		** (lgcairo:68503): DEBUG: glg_line_graph_master_draw(exited)

      ...
		
		** (lgcairo:51717): DEBUG: ===> glg_line_graph_destroy(enter)
		** (lgcairo:51717): DEBUG: ===> glg_line_graph_data_series_remove_all()
		** (lgcairo:51717): DEBUG:   ==>DataSeriesRemoveAll: number removed=5
		** (lgcairo:51717): DEBUG: glg_line_graph_destroy(exit)
		** (lgcairo:51717): DEBUG: ===> glg_line_graph_destroy(enter)
		** (lgcairo:51717): DEBUG: glg_line_graph_destroy(exit)

END DEBUG LOG:
 	
