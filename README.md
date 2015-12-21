smash, a terminal emulator experiment  
Copyright 2015 Evan Martin.

All day long I work using a shell under a terminal emulator -- a
combination of software that hasn't changed much in 30 years.  Why?  I
use the command line because it is powerful and expressive and because
all the other tools I use work well with it.  But at the same time it
seems we ought be able to improve on the interface while keeping the
fundamental idea.

Smash is an attempt to improve the textual user interface while
preserving the the good parts.  Reasonable people can disagree over
what exactly that means.  Ideas are welcome.

## Features

Smash integrates the shell, terminal emulator, and GUI into a single
program so they can work as one.  This allows Smash to:

* smooth scroll the display, even when running Smash-unaware programs
  like `less`.

* tab completion is a popup, yet it reuses the existing bash completion logic.

## Non-goals

Smash is *not*:

* a new shell scripting language; nobody needs yet another one of those.

* a replacement for the Unix suite of commands; my fingers remember "ls".

* intended for inexperienced users; while there is interesting work
  being done in newbie-friendly shells, it is not a problem that I
  am interested in solving.

## Future work

Some other things I plan to implement:

* concurrently running commands should not interleave their output.

* job control can be exposed visually.

Here are some other ideas that I am considering:

* when you ssh somewhere, all typing ought to be handled locally; you
  only should need to round-trip to the remote host for tab completion
  and command execution.

* perhaps you can introduce a type-system-like structure to shell
  pipelines, in that text streams can sometimes be viewed as arrays of
  lines and lines can sometimes be viewed as arrays of columns.
