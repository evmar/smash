export type uint8 = number;
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
    const arr: T[] = new Array(len);
    for (let i = 0; i < len; i++) {
      arr[i] = elem();
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
      id: this.readInt(),
      cwd: this.readString(),
      input: this.readString(),
      pos: this.readInt(),
    });
  }
  readCompleteResponse(): CompleteResponse {
    return new CompleteResponse({
      id: this.readInt(),
      error: this.readString(),
      pos: this.readInt(),
      completions: this.readArray(() => this.readString()),
    });
  }
  readRunRequest(): RunRequest {
    return new RunRequest({
      cell: this.readInt(),
      cwd: this.readString(),
      argv: this.readArray(() => this.readString()),
    });
  }
  readKeyEvent(): KeyEvent {
    return new KeyEvent({
      cell: this.readInt(),
      keys: this.readString(),
    });
  }
  readRowSpans(): RowSpans {
    return new RowSpans({
      row: this.readInt(),
      spans: this.readArray(() => this.readSpan()),
    });
  }
  readSpan(): Span {
    return new Span({
      attr: this.readInt(),
      text: this.readString(),
    });
  }
  readCursor(): Cursor {
    return new Cursor({
      row: this.readInt(),
      col: this.readInt(),
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
      exitCode: this.readInt(),
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
      cell: this.readInt(),
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
    this.writeInt(msg.cell);
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
