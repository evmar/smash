#include <gtk/gtk.h>

#include "smashgtk.h"

extern "C" {

#include "_cgo_export.h"

void smash_gtk_init(void) {
  // TODO: command-line params?
  gtk_init(NULL, NULL);
}

static void draw(GtkWidget* widget, cairo_t* cr, gpointer data) {
  smashGoDraw(data, cr);
}

static void key(GtkWidget* widget, GdkEventKey* event, gpointer data) {
  smashGoKey(data, event);
}

GtkWidget* smash_gtk_new_window(void* delegate, int toplevel) {
  GtkWidget* win = gtk_window_new(toplevel ? GTK_WINDOW_TOPLEVEL
                                  : GTK_WINDOW_POPUP);
  gtk_window_set_title(GTK_WINDOW(win), "smash");
  gtk_widget_set_app_paintable(win, TRUE);
  g_signal_connect(win, "draw", G_CALLBACK(draw), delegate);
  g_signal_connect(win, "key-press-event", G_CALLBACK(key), delegate);

  gtk_widget_realize(win);

  // GdkGeometry geo = {};
  // geo.width_inc = term_->font_metrics_.width;
  // geo.height_inc = term_->font_metrics_.height;
  // gtk_window_set_geometry_hints(GTK_WINDOW(win_), nullptr,
  //                               &geo, GDK_HINT_RESIZE_INC);
  // gtk_window_set_default_size(GTK_WINDOW(win_),
  //                             term_->width_, term_->height_);

  return win;
}

int smash_idle_cb(void* data) {
  return smashGoIdle(data);
}

static gboolean tick(GtkWidget* widget, GdkFrameClock* clock, gpointer data) {
  gboolean more = smashGoTick(data) != 0;
  gtk_widget_queue_draw(widget);
  return more;
}
void smash_start_ticks(void* data, GtkWidget* widget) {
  gtk_widget_add_tick_callback(widget, tick, data, NULL);
}

}  // extern "C"
