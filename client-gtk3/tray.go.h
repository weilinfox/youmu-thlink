
#ifndef __TRAY_GO_H__
#define __TRAY_GO_H__

#pragma once

#include <gdk/gdk.h>
#include <gtk/gtk.h>

const char * thlink_client_gtk_tray_xpm[] = {
"32 32 3 1",
" 	c None",
".	c #FFFFFF",
">	c #000000",
"     ......................     ",
"   ..........................   ",
"  ............................  ",
" .............................. ",
" .............................. ",
"..>>>>>>>>>.>>.......>.>>>>>>...",
"..>>>>>>>>>.>>.......>.>>>>.....",
".......>>...>>>>>>>>>>..>.......",
".......>>...>>.......>..>.......",
".....>>>>...>>.......>..>>......",
".....>>>>...............>.......",
".......>>.>>>>>>>>>>>>>.>.......",
".......>>.>>............>.......",
".......>>...............>.......",
".......>>...............>.......",
".......>>...............>.......",
".......>>...............>.......",
".......>>..........>>>..>.......",
".......>>...............>.......",
".......>>...............>.......",
".......>>.......>>>>....>.......",
".......>>...............>.......",
".......>>...............>.......",
".......>>...............>.......",
".......>>.....>>>>>.....>.......",
".......>>.....>>>>>.....>.......",
".......>>.......................",
" .............................. ",
" .............................. ",
"  ............................  ",
"   ..............>>>>>>......   ",
"     ............>>>>>>....     "};

GtkStatusIcon *_icon;

void _activate_signal_callback( GObject* trayIcon, gpointer window )
{
	if (gtk_widget_get_visible(GTK_WIDGET(window))) {
		gtk_widget_hide(GTK_WIDGET(window));
	} else {
		gtk_widget_show_all(GTK_WIDGET(window));
	}
}

void _status_icon_activate_signal_connect( GtkStatusIcon *trayIcon, gpointer window )
{
	g_signal_connect(GTK_STATUS_ICON (trayIcon), "activate",  G_CALLBACK(_activate_signal_callback), GTK_WIDGET(window));
}

void status_icon_setup( gpointer window )
{
    if (_icon != NULL) {
        return;
    }
	_icon = gtk_status_icon_new_from_pixbuf(gdk_pixbuf_new_from_xpm_data(thlink_client_gtk_tray_xpm));
	_status_icon_activate_signal_connect( _icon, window );
	gtk_status_icon_set_visible( _icon, TRUE );

    // where are name and tooltip shown ?
	gtk_status_icon_set_name( _icon, "ThLink" );
	gtk_status_icon_set_tooltip_text(_icon, "ThLink client");
	gtk_status_icon_set_has_tooltip( _icon, TRUE );
}

void status_icon_title_set( const char *text )
{
    gtk_status_icon_set_title(_icon, text);
}

#endif
