export type uint8 = number;
export type ClientMessage =
  | { tag: 'CompleteRequest'; val: CompleteRequest }
  | { tag: 'RunRequest'; val: RunRequest }
  | { tag: 'KeyEvent'; val: KeyEvent };
export interface CompleteRequest {
  id: number;
  cwd: string;
  input: string;
  pos: number;
}
export interface CompleteResponse {
  id: number;
  error: string;
  pos: number;
  completions: string[];
}
export interface RunRequest {
  cell: number;
  cwd: string;
  argv: string[];
}
export interface KeyEvent {
  cell: number;
  keys: string;
}
export interface RowSpans {
  row: number;
  spans: Span[];
}
export interface Span {
  attr: number;
  text: string;
}
export interface Cursor {
  row: number;
  col: number;
  hidden: boolean;
}
export interface TermUpdate {
  rows: RowSpans[];
  cursor: Cursor;
}
export interface Pair {
  key: string;
  val: string;
}
export interface Hello {
  alias: Pair[];
  env: Pair[];
}
export interface CmdError {
  error: string;
}
export interface Exit {
  exitCode: number;
}
export type Output =
  | { tag: 'CmdError'; val: CmdError }
  | { tag: 'TermUpdate'; val: TermUpdate }
  | { tag: 'Exit'; val: Exit };
export interface CellOutput {
  cell: number;
  output: Output;
}
export type ServerMsg =
  | { tag: 'Hello'; val: Hello }
  | { tag: 'CompleteResponse'; val: CompleteResponse }
  | { tag: 'CellOutput'; val: CellOutput };
export class Reader {
  private ofs = 0;
  constructor(readonly view: DataView) {}

  private readUint8(): number {
    return this.view.getUint8(this.ofs++);
  }

  private readInt(): number {
    let val = 0;
    let shift = 0;
    for (;;) {
      const b = this.readUint8();
      val |= (b & 0x7f) << shift;
      if ((b & 0x80) === 0) break;
      shift += 7;
    }
    return val;
  }

  private readBoolean(): boolean {
    return this.readUint8() !== 0;
  }

  private readBytes(): DataView {
    const len = this.readInt();
    const slice = new DataView(this.view.buffer, this.ofs, len);
    this.ofs += len;
    return slice;
  }

  private readString(): string {
    const bytes = this.readBytes();
    return new TextDecoder().decode(bytes);
  }

  private readArray<T>(elem: () => T): T[] {
    const len = this.readInt();
    const arr: T[] = [];
    for (let i = 0; i < len; i++) {
      arr.push(elem());
    }
    return arr;
  }
  readClientMessage(): ClientMessage {
    switch (this.readUint8()) {
      case 1:
        return { tag: 'CompleteRequest', val: this.readCompleteRequest() };
      case 2:
        return { tag: 'RunRequest', val: this.readRunRequest() };
      case 3:
        return { tag: 'KeyEvent', val: this.readKeyEvent() };
      default:
        throw new Error('parse error');
    }
  }
  readCompleteRequest(): CompleteRequest {
    return {
      id: this.readInt(),
      cwd: this.readString(),
      input: this.readString(),
      pos: this.readInt(),
    };
  }
  readCompleteResponse(): CompleteResponse {
    return {
      id: this.readInt(),
      error: this.readString(),
      pos: this.readInt(),
      completions: this.readArray(() => this.readString()),
    };
  }
  readRunRequest(): RunRequest {
    return {
      cell: this.readInt(),
      cwd: this.readString(),
      argv: this.readArray(() => this.readString()),
    };
  }
  readKeyEvent(): KeyEvent {
    return {
      cell: this.readInt(),
      keys: this.readString(),
    };
  }
  readRowSpans(): RowSpans {
    return {
      row: this.readInt(),
      spans: this.readArray(() => this.readSpan()),
    };
  }
  readSpan(): Span {
    return {
      attr: this.readInt(),
      text: this.readString(),
    };
  }
  readCursor(): Cursor {
    return {
      row: this.readInt(),
      col: this.readInt(),
      hidden: this.readBoolean(),
    };
  }
  readTermUpdate(): TermUpdate {
    return {
      rows: this.readArray(() => this.readRowSpans()),
      cursor: this.readCursor(),
    };
  }
  readPair(): Pair {
    return {
      key: this.readString(),
      val: this.readString(),
    };
  }
  readHello(): Hello {
    return {
      alias: this.readArray(() => this.readPair()),
      env: this.readArray(() => this.readPair()),
    };
  }
  readCmdError(): CmdError {
    return {
      error: this.readString(),
    };
  }
  readExit(): Exit {
    return {
      exitCode: this.readInt(),
    };
  }
  readOutput(): Output {
    switch (this.readUint8()) {
      case 1:
        return { tag: 'CmdError', val: this.readCmdError() };
      case 2:
        return { tag: 'TermUpdate', val: this.readTermUpdate() };
      case 3:
        return { tag: 'Exit', val: this.readExit() };
      default:
        throw new Error('parse error');
    }
  }
  readCellOutput(): CellOutput {
    return {
      cell: this.readInt(),
      output: this.readOutput(),
    };
  }
  readServerMsg(): ServerMsg {
    switch (this.readUint8()) {
      case 1:
        return { tag: 'Hello', val: this.readHello() };
      case 2:
        return { tag: 'CompleteResponse', val: this.readCompleteResponse() };
      case 3:
        return { tag: 'CellOutput', val: this.readCellOutput() };
      default:
        throw new Error('parse error');
    }
  }
}
export class Writer {
  public ofs = 0;
  public buf = new Uint8Array();
  writeBoolean(val: boolean) {
    this.writeUint8(val ? 1 : 0);
  }
  writeUint8(val: number) {
    if (val > 0xff) throw new Error('overflow');
    this.buf[this.ofs++] = val;
  }
  writeInt(val: number) {
    if (val < 0) throw new Error('negative');
    for (;;) {
      const b = val & 0x7f;
      val = val >> 7;
      if (val === 0) {
        this.writeUint8(b);
        return;
      }
      this.writeUint8(b | 0x80);
    }
  }
  writeString(str: string) {
    this.writeInt(str.length);
    for (let i = 0; i < str.length; i++) {
      this.buf[this.ofs++] = str.charCodeAt(i);
    }
  }
  writeArray<T>(arr: T[], f: (t: T) => void) {
    this.writeInt(arr.length);
    for (const elem of arr) {
      f(elem);
    }
  }
  writeClientMessage(msg: ClientMessage) {
    switch (msg.tag) {
      case 'CompleteRequest':
        this.writeUint8(1);
        this.writeCompleteRequest(msg.val);
        break;
      case 'RunRequest':
        this.writeUint8(2);
        this.writeRunRequest(msg.val);
        break;
      case 'KeyEvent':
        this.writeUint8(3);
        this.writeKeyEvent(msg.val);
        break;
    }
  }
  writeCompleteRequest(msg: CompleteRequest) {
    this.writeInt(msg.id);
    this.writeString(msg.cwd);
    this.writeString(msg.input);
    this.writeInt(msg.pos);
  }
  writeCompleteResponse(msg: CompleteResponse) {
    this.writeInt(msg.id);
    this.writeString(msg.error);
    this.writeInt(msg.pos);
    this.writeArray(msg.completions, (val) => {
      this.writeString(val);
    });
  }
  writeRunRequest(msg: RunRequest) {
    this.writeInt(msg.cell);
    this.writeString(msg.cwd);
    this.writeArray(msg.argv, (val) => {
      this.writeString(val);
    });
  }
  writeKeyEvent(msg: KeyEvent) {
    this.writeInt(msg.cell);
    this.writeString(msg.keys);
  }
  writeRowSpans(msg: RowSpans) {
    this.writeInt(msg.row);
    this.writeArray(msg.spans, (val) => {
      this.writeSpan(val);
    });
  }
  writeSpan(msg: Span) {
    this.writeInt(msg.attr);
    this.writeString(msg.text);
  }
  writeCursor(msg: Cursor) {
    this.writeInt(msg.row);
    this.writeInt(msg.col);
    this.writeBoolean(msg.hidden);
  }
  writeTermUpdate(msg: TermUpdate) {
    this.writeArray(msg.rows, (val) => {
      this.writeRowSpans(val);
    });
    this.writeCursor(msg.cursor);
  }
  writePair(msg: Pair) {
    this.writeString(msg.key);
    this.writeString(msg.val);
  }
  writeHello(msg: Hello) {
    this.writeArray(msg.alias, (val) => {
      this.writePair(val);
    });
    this.writeArray(msg.env, (val) => {
      this.writePair(val);
    });
  }
  writeCmdError(msg: CmdError) {
    this.writeString(msg.error);
  }
  writeExit(msg: Exit) {
    this.writeInt(msg.exitCode);
  }
  writeOutput(msg: Output) {
    switch (msg.tag) {
      case 'CmdError':
        this.writeUint8(1);
        this.writeCmdError(msg.val);
        break;
      case 'TermUpdate':
        this.writeUint8(2);
        this.writeTermUpdate(msg.val);
        break;
      case 'Exit':
        this.writeUint8(3);
        this.writeExit(msg.val);
        break;
    }
  }
  writeCellOutput(msg: CellOutput) {
    this.writeInt(msg.cell);
    this.writeOutput(msg.output);
  }
  writeServerMsg(msg: ServerMsg) {
    switch (msg.tag) {
      case 'Hello':
        this.writeUint8(1);
        this.writeHello(msg.val);
        break;
      case 'CompleteResponse':
        this.writeUint8(2);
        this.writeCompleteResponse(msg.val);
        break;
      case 'CellOutput':
        this.writeUint8(3);
        this.writeCellOutput(msg.val);
        break;
    }
  }
}
