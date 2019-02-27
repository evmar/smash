smash, a terminal emulator experiment

All day long I work using a shell under a terminal emulator -- a
combination of software that hasn't changed much in 30 years. Why? I
use the command line because it is powerful and expressive and because
all the other tools I use work well with it. But at the same time it
seems we ought be able to improve on the interface while keeping the
fundamental idea.

Smash is an attempt to improve the textual user interface while
preserving the the good parts. Reasonable people can disagree over
what exactly that means. If this whole idea makes you upset please
see the first question in the [FAQ](FAQ.md).

## Non-goals

Smash is _not_:

- a new shell scripting language; nobody needs yet another one of those.

- a replacement for the Unix suite of commands; my fingers remember "ls".

- intended for inexperienced users; while there is interesting work
  being done in newbie-friendly shells, it is not a problem that I
  am interested in solving.

## Design goals

- Zero-latency interactions even when connected to a remote host, using
  a client-server architecture where the user interacts with a local
  process (e.g. native UI) and the commands execute remotely (e.g. via
  ssh). Compare to [mosh](https://mosh.org/), though mosh is both
  more primitive and more sophisticated.

- Native UI for local interactions like text editing and scrollback.

## Related work

The smash wiki has a collection of links to [terminal emulator
experiments](https://github.com/evmar/smash/wiki/Related-projects).
There are many interesting ideas in there!
