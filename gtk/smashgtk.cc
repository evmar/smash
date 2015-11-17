#include <gtk/gtk.h>

#include "smashgtk.h"

static void activate(GtkApplication* app, gpointer data) {
  GtkWidget* win = gtk_application_window_new(app);
  gtk_window_set_title(GTK_WINDOW(win), "smash");
  gtk_widget_set_app_paintable(win, TRUE);
  //g_signal_connect(win, "draw", G_CALLBACK(draw), NULL);
  //g_signal_connect(win, "key-press-event", G_CALLBACK(key), NULL);

  /*
 gtk_widget_realize(win);
 cairo_t* cr = gdk_cairo_create(gtk_widget_get_window(win));
 term_->measure(cr.get());
 GdkGeometry geo = {};
 geo.width_inc = term_->font_metrics_.width;
 geo.height_inc = term_->font_metrics_.height;
 gtk_window_set_geometry_hints(GTK_WINDOW(win_), nullptr,
                               &geo, GDK_HINT_RESIZE_INC);
 gtk_window_set_default_size(GTK_WINDOW(win_),
                              term_->width_, term_->height_);
  */
  gtk_widget_show(win);
}

int smash_gtk_main(int argc, char** argv) {
  GtkApplication* app = gtk_application_new("org.neugierig.smash",
                                            G_APPLICATION_FLAGS_NONE);
  g_signal_connect(app, "activate",
                   G_CALLBACK(activate), NULL);

  return g_application_run(G_APPLICATION(app), argc, argv);
}

extern "C" {
void run(void) {
  smash_gtk_main(0, NULL);
}
}
