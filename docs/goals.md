# smash: Project goals

## Non-goals

Smash is _not_:

- A new shell scripting language; nobody needs yet another one of those.

- A replacement for the Unix suite of commands; my fingers remember "ls".

- Intended for inexperienced users; while there is interesting work being done
  in newbie-friendly shells, it is not a problem that I am interested in
  solving.

## Design goals

- Zero-latency interactions even when connected to a remote host, using a
  client-server architecture where the user interacts with a local process (e.g.
  native UI) and the commands execute remotely (e.g. via ssh). Compare to
  [mosh](https://mosh.org/), though mosh is both more primitive and more
  sophisticated.

- Native UI for local interactions like text editing and scrollback.

## Related work

The smash wiki has a collection of links to
[terminal emulator experiments](https://github.com/evmar/smash/wiki/Related-projects).
There are many interesting ideas in there!
