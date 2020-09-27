# Design

Smash merges the shell, terminal emulator, and client-side app. Smash tries to
not replace the shell semantics: commands are still the unix commands, pipelines
are still pipelines, tab completion still comes from bash.

The UX goals of smash are:

- No latency on "client-side" interactions like moving the cursor within the
  prompt, even when connected to a remote host.
- Support the mouse where appropriate.
- Window management, e.g. tabs.
- Native interactions for job control and executed subcommands; e.g. native
  popup for tab completion, native scrollbars for scrollback, and integrations
  like "run this command in a new tab". Concurrently running commands never
  interleave their output.
- Terminal state is persisted server-side, so disconnecting and reconnecting
  resumes your session.
- Preserve most of the keystrokes used in an ordinary bash.

Together, these goals produced this design:

- The client implments the shell prompt, keyboard interactions, windowing etc.
- The server executes commands and implements terminal emulation. (Otherwise you
  couldn't restore the screen state when the client reconnects.)
- The two communicate via a custom protocol (command to execute in one
  direction, render output in the other).

The smash server process interprets the terminal output, which means it can
expose that output back to the shell.

## Deferred work

I am definitely interested in revisiting the semantics of commands, like how
pipelines work, but I am attempting to limit scope to something tractable to
complete. See the [related work](related.md) for some inspiration.
