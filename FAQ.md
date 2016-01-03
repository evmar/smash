# Frequently Asked Questions
        
## Why is [thing X] different than the way it normally works?

It may be the case that everything as it currently is is already
perfect and should not be changed.  However, many good new ideas are
only obvious in retrospect and cannot be discovered in an environment
where any deviation from the status quo is immediately shot down.

The purpose of this project is to explore new directions for the
command line user interface.  If you think that the way things
currently work cannot be improved on, you should just use one of the
existing terminal emulators and keep your criticism to yourself.

## Why is Smash a GUI app and not just a console app?

* Smash scrolls lines on sub-character pixel boundaries.
* Smash pops up new windows as part of tab completion.

## Why not just run bash/zsh within the terminal emulator?

## Why does Smash depend on GTK?

Smash doesn't really use GTK -- it just needs to bring up a window and
draw to it.  Importantly, GTK implements the logic to synchronize with
an X compositor so that animations are smooth.

The GUI code is abstracted to talk to an interface.  There is a
bitrotted (nonworking) plain-xlib implementation (build with `-tags
xlib`) that lacks the compositor bits.
