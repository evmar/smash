#include <gtk/gtk.h>

#ifdef __cplusplus
extern "C" {
#endif

void smash_gtk_init(void);
GtkWidget* smash_gtk_new_window(void* delegate);
int smash_idle_cb(void* data);
void smash_start_ticks(void* data, GtkWidget* widget);

#ifdef __cplusplus
}
#endif
