# Frequently Asked Questions
        
## Why is [thing X] different than the way it normally works?

The purpose of this project is to explore new directions for the
command line user interface.  If you think that the way things
currently work cannot be improved on, you should use one of the
existing terminal emulators.

## Why is Smash a GUI app and not just a console app?

* Smash scrolls lines on sub-character pixel boundaries.
* Smash pops up new windows as part of tab completion.

## Why not just run bash/zsh within the terminal emulator?

## Why does Smash depend on GTK?

Smash doesn't really use GTK, it just only needs to bring up a window
and draw to it.  That code is abstracted to talk to an interface.
There's even a plain-xlib implementation (build with `-tags xlib`) but
it's bitrotted a bit.  Importantly, GTK implements the logic to
synchronize with an X compositor so that animations are smooth.
