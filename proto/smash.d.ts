type uint8 = number;
type uint16 = number;

/** Message from client to server. */
type ClientMessage = CompleteRequest | RunRequest | KeyEvent;

/** Request to complete a partial command-line input. */
interface CompleteRequest {
  id: uint16;
  cwd: string;
  input: string;
  pos: uint16;
}

/** Response to a CompleteRequest. */
interface CompleteResponse {
  id: uint16;
  error: string;
  pos: uint16;
  completions: string[];
}

/** Request to spawn a command. */
interface RunRequest {
  cell: uint16;
  cwd: string;
  argv: string[];
}

/** Keystroke sent to running command. */
interface KeyEvent {
  cell: uint16;
  keys: string;
}

interface RowSpans {
  row: uint16;
  spans: Span[];
}
interface Span {
  attr: uint16;
  text: string;
}
interface Cursor {
  row: uint16;
  col: uint16;
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
  exitCode: uint16;
}
type Output = CmdError | TermUpdate | Exit;

/** Message from server to client about a running subprocess. */
interface CellOutput {
  cell: uint16;
  output: Output;
}

type ServerMsg = Hello | CompleteResponse | CellOutput;
