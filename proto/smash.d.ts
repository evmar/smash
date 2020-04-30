/**
 * This file defines the smash client<->server protocol.
 * Note that the types as written here are not actually used directly
 * in smash, but rather more complex types are derived from them.
 * See README.md.
 */

type int = number;
type uint8 = number;

/** Message from client to server. */
type ClientMessage = CompleteRequest | RunRequest | KeyEvent;

/** Request to complete a partial command-line input. */
interface CompleteRequest {
  id: int;
  cwd: string;
  input: string;
  pos: int;
}

/** Response to a CompleteRequest. */
interface CompleteResponse {
  id: int;
  error: string;
  pos: int;
  completions: string[];
}

/** Request to spawn a command. */
interface RunRequest {
  cell: int;
  cwd: string;
  argv: string[];
}

/** Keystroke sent to running command. */
interface KeyEvent {
  cell: int;
  keys: string;
}

interface RowSpans {
  row: int;
  spans: Span[];
}
interface Span {
  attr: int;
  text: string;
}
interface Cursor {
  row: int;
  col: int;
  hidden: boolean;
}

/** Termial update, server -> client. */
interface TermUpdate {
  /** Updates to specific rows of output. */
  rows: RowSpans[];
  /** Cursor position. */
  cursor: Cursor;
}

interface Pair {
  key: string;
  val: string;
}

/** Message from server to client on connection. */
interface Hello {
  /** Command aliases, from alias name to expansion. */
  alias: Pair[];

  /** Environment variables. */
  env: Pair[];

  // TODO: running cells and their state.
}

interface CmdError {
  error: string;
}
interface Exit {
  exitCode: int;
}
type Output = CmdError | TermUpdate | Exit;

/** Message from server to client about a running subprocess. */
interface CellOutput {
  cell: int;
  output: Output;
}

type ServerMsg = Hello | CompleteResponse | CellOutput;
