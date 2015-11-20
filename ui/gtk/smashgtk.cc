#include <gtk/gtk.h>

#include "smashgtk.h"

namespace {

extern "C" void smashGoDraw(void*, void*);
void draw(GtkWidget* widget, cairo_t* cr, gpointer data) {
  smashGoDraw(data, cr);
}

extern "C" void smashGoKey(void*, void*);
void key(GtkWidget* widget, GdkEventKey* event, gpointer data) {
  smashGoKey(data, event);
}

extern "C" int smashGoTick(void*);
gboolean tick(GtkWidget* widget, GdkFrameClock* clock, gpointer data) {
  gboolean more = smashGoTick(data) != 0;
  gtk_widget_queue_draw(widget);
  return more;
}

}  // anonymous namespace

extern "C" {

void smash_gtk_init(void) {
  gtk_init(NULL, NULL);
}

SmashWin* smash_gtk_new_window(SmashWinDelegate* delegate) {
  GtkWidget* win = gtk_window_new(GTK_WINDOW_TOPLEVEL);
  gtk_window_set_title(GTK_WINDOW(win), "smash");
  gtk_widget_set_app_paintable(win, TRUE);
  g_signal_connect(win, "draw", G_CALLBACK(draw), delegate);
  g_signal_connect(win, "key-press-event", G_CALLBACK(key), delegate);

  // gtk_widget_realize(win);
  // cairo_t* cr = gdk_cairo_create(gtk_widget_get_window(win));
  // term_->measure(cr.get());
  // GdkGeometry geo = {};
  // geo.width_inc = term_->font_metrics_.width;
  // geo.height_inc = term_->font_metrics_.height;
  // gtk_window_set_geometry_hints(GTK_WINDOW(win_), nullptr,
  //                               &geo, GDK_HINT_RESIZE_INC);
  // gtk_window_set_default_size(GTK_WINDOW(win_),
  //                             term_->width_, term_->height_);
  // cairo_destroy(cr);
  gtk_window_set_default_size(GTK_WINDOW(win),
                              640, 400);
 
  gtk_widget_show(win);
  return win;
}

extern "C" int smashGoIdle(void*);
int smash_idle_cb(void* data) {
  return smashGoIdle(data);
}

void smash_start_ticks(void* data, GtkWidget* widget) {
  gtk_widget_add_tick_callback(widget, tick, data, NULL);
}

}
