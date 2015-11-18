
#ifdef __cplusplus
extern "C" {
#endif

void smash_gtk_init(void);

typedef void SmashWin;
typedef void SmashWinDelegate;
SmashWin* smash_gtk_new_window(SmashWinDelegate* delegate);

int smash_idle_cb(void* data);

#ifdef __cplusplus
}
#endif
