export type uint8 = number;
export type uint16 = number;
export class ClientMessage {
  constructor(public alt: CompleteRequest | RunRequest | KeyEvent) {}
}
export class CompleteRequest {
  id!: number;
  cwd!: string;
  input!: string;
  pos!: number;
  constructor(fields: CompleteRequest) {
    Object.assign(this, fields);
  }
}
export class CompleteResponse {
  id!: number;
  error!: string;
  pos!: number;
  completions!: string[];
  constructor(fields: CompleteResponse) {
    Object.assign(this, fields);
  }
}
export class RunRequest {
  cell!: number;
  cwd!: string;
  argv!: string[];
  constructor(fields: RunRequest) {
    Object.assign(this, fields);
  }
}
export class KeyEvent {
  cell!: number;
  keys!: string;
  constructor(fields: KeyEvent) {
    Object.assign(this, fields);
  }
}
export class RowSpans {
  row!: number;
  spans!: Span[];
  constructor(fields: RowSpans) {
    Object.assign(this, fields);
  }
}
export class Span {
  attr!: number;
  text!: string;
  constructor(fields: Span) {
    Object.assign(this, fields);
  }
}
export class Cursor {
  row!: number;
  col!: number;
  hidden!: boolean;
  constructor(fields: Cursor) {
    Object.assign(this, fields);
  }
}
export class TermUpdate {
  rows!: RowSpans[];
  cursor!: Cursor;
  constructor(fields: TermUpdate) {
    Object.assign(this, fields);
  }
}
export class Pair {
  key!: string;
  val!: string;
  constructor(fields: Pair) {
    Object.assign(this, fields);
  }
}
export class Hello {
  alias!: Pair[];
  env!: Pair[];
  constructor(fields: Hello) {
    Object.assign(this, fields);
  }
}
export class CmdError {
  error!: string;
  constructor(fields: CmdError) {
    Object.assign(this, fields);
  }
}
export class Exit {
  exitCode!: number;
  constructor(fields: Exit) {
    Object.assign(this, fields);
  }
}
export class Output {
  constructor(public alt: CmdError | TermUpdate | Exit) {}
}
export class CellOutput {
  cell!: number;
  output!: Output;
  constructor(fields: CellOutput) {
    Object.assign(this, fields);
  }
}
export class ServerMsg {
  constructor(public alt: Hello | CompleteResponse | CellOutput) {}
}
export class Reader {
  private ofs = 0;
  constructor(readonly view: DataView) {}

  private readUint8(): number {
    return this.view.getUint8(this.ofs++);
  }

  private readUint16(): number {
    const val = this.view.getUint16(this.ofs);
    this.ofs += 2;
    return val;
  }

  private readBoolean(): boolean {
    return this.readUint8() !== 0;
  }

  private readBytes(): DataView {
    const len = this.readUint16();
    const slice = new DataView(this.view.buffer, this.ofs, len);
    this.ofs += len;
    return slice;
  }

  private readString(): string {
    const bytes = this.readBytes();
    return new TextDecoder().decode(bytes);
  }

  private readArray<T>(elem: () => T): T[] {
    const len = this.readUint8();
    const arr: T[] = [];
    for (let i = 0; i < len; i++) {
      arr.push(elem());
    }
    return arr;
  }
  readClientMessage(): ClientMessage {
    switch (this.readUint8()) {
      case 1:
        return new ClientMessage(this.readCompleteRequest());
      case 2:
        return new ClientMessage(this.readRunRequest());
      case 3:
        return new ClientMessage(this.readKeyEvent());
      default:
        throw new Error('parse error');
    }
  }
  readCompleteRequest(): CompleteRequest {
    return new CompleteRequest({
      id: this.readUint16(),
      cwd: this.readString(),
      input: this.readString(),
      pos: this.readUint16(),
    });
  }
  readCompleteResponse(): CompleteResponse {
    return new CompleteResponse({
      id: this.readUint16(),
      error: this.readString(),
      pos: this.readUint16(),
      completions: this.readArray(() => this.readString()),
    });
  }
  readRunRequest(): RunRequest {
    return new RunRequest({
      cell: this.readUint16(),
      cwd: this.readString(),
      argv: this.readArray(() => this.readString()),
    });
  }
  readKeyEvent(): KeyEvent {
    return new KeyEvent({
      cell: this.readUint16(),
      keys: this.readString(),
    });
  }
  readRowSpans(): RowSpans {
    return new RowSpans({
      row: this.readUint16(),
      spans: this.readArray(() => this.readSpan()),
    });
  }
  readSpan(): Span {
    return new Span({
      attr: this.readUint16(),
      text: this.readString(),
    });
  }
  readCursor(): Cursor {
    return new Cursor({
      row: this.readUint16(),
      col: this.readUint16(),
      hidden: this.readBoolean(),
    });
  }
  readTermUpdate(): TermUpdate {
    return new TermUpdate({
      rows: this.readArray(() => this.readRowSpans()),
      cursor: this.readCursor(),
    });
  }
  readPair(): Pair {
    return new Pair({
      key: this.readString(),
      val: this.readString(),
    });
  }
  readHello(): Hello {
    return new Hello({
      alias: this.readArray(() => this.readPair()),
      env: this.readArray(() => this.readPair()),
    });
  }
  readCmdError(): CmdError {
    return new CmdError({
      error: this.readString(),
    });
  }
  readExit(): Exit {
    return new Exit({
      exitCode: this.readUint16(),
    });
  }
  readOutput(): Output {
    switch (this.readUint8()) {
      case 1:
        return new Output(this.readCmdError());
      case 2:
        return new Output(this.readTermUpdate());
      case 3:
        return new Output(this.readExit());
      default:
        throw new Error('parse error');
    }
  }
  readCellOutput(): CellOutput {
    return new CellOutput({
      cell: this.readUint16(),
      output: this.readOutput(),
    });
  }
  readServerMsg(): ServerMsg {
    switch (this.readUint8()) {
      case 1:
        return new ServerMsg(this.readHello());
      case 2:
        return new ServerMsg(this.readCompleteResponse());
      case 3:
        return new ServerMsg(this.readCellOutput());
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
  writeUint16(val: number) {
    if (val > 0xffff) throw new Error('overflow');
    this.buf[this.ofs++] = (val & 0xff00) >> 8;
    this.buf[this.ofs++] = val & 0xff;
  }
  writeString(str: string) {
    this.writeUint16(str.length);
    for (let i = 0; i < str.length; i++) {
      this.buf[this.ofs++] = str.charCodeAt(i);
    }
  }
  writeArray<T>(arr: T[], f: (t: T) => void) {
    this.writeUint16(arr.length);
    for (const elem of arr) {
      f(elem);
    }
  }
  writeClientMessage(msg: ClientMessage) {
    if (msg.alt instanceof CompleteRequest) {
      this.writeUint8(1);
      this.writeCompleteRequest(msg.alt);
    } else if (msg.alt instanceof RunRequest) {
      this.writeUint8(2);
      this.writeRunRequest(msg.alt);
    } else if (msg.alt instanceof KeyEvent) {
      this.writeUint8(3);
      this.writeKeyEvent(msg.alt);
    } else {
      throw new Error('unhandled case');
    }
  }
  writeCompleteRequest(msg: CompleteRequest) {
    this.writeUint16(msg.id);
    this.writeString(msg.cwd);
    this.writeString(msg.input);
    this.writeUint16(msg.pos);
  }
  writeCompleteResponse(msg: CompleteResponse) {
    this.writeUint16(msg.id);
    this.writeString(msg.error);
    this.writeUint16(msg.pos);
    this.writeArray(msg.completions, (val) => {
      this.writeString(val);
    });
  }
  writeRunRequest(msg: RunRequest) {
    this.writeUint16(msg.cell);
    this.writeString(msg.cwd);
    this.writeArray(msg.argv, (val) => {
      this.writeString(val);
    });
  }
  writeKeyEvent(msg: KeyEvent) {
    this.writeUint16(msg.cell);
    this.writeString(msg.keys);
  }
  writeRowSpans(msg: RowSpans) {
    this.writeUint16(msg.row);
    this.writeArray(msg.spans, (val) => {
      this.writeSpan(val);
    });
  }
  writeSpan(msg: Span) {
    this.writeUint16(msg.attr);
    this.writeString(msg.text);
  }
  writeCursor(msg: Cursor) {
    this.writeUint16(msg.row);
    this.writeUint16(msg.col);
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
    this.writeUint16(msg.exitCode);
  }
  writeOutput(msg: Output) {
    if (msg.alt instanceof CmdError) {
      this.writeUint8(1);
      this.writeCmdError(msg.alt);
    } else if (msg.alt instanceof TermUpdate) {
      this.writeUint8(2);
      this.writeTermUpdate(msg.alt);
    } else if (msg.alt instanceof Exit) {
      this.writeUint8(3);
      this.writeExit(msg.alt);
    } else {
      throw new Error('unhandled case');
    }
  }
  writeCellOutput(msg: CellOutput) {
    this.writeUint16(msg.cell);
    this.writeOutput(msg.output);
  }
  writeServerMsg(msg: ServerMsg) {
    if (msg.alt instanceof Hello) {
      this.writeUint8(1);
      this.writeHello(msg.alt);
    } else if (msg.alt instanceof CompleteResponse) {
      this.writeUint8(2);
      this.writeCompleteResponse(msg.alt);
    } else if (msg.alt instanceof CellOutput) {
      this.writeUint8(3);
      this.writeCellOutput(msg.alt);
    } else {
      throw new Error('unhandled case');
    }
  }
}
