import * as pb from './smash_pb';
import * as sh from './shell';

let ws: WebSocket | null = null;
let shell = new sh.Shell();
let cellStack: CellStack;

function html(tagName: string, attr: { [key: string]: {} } = {}) {
  const tag = document.createElement(tagName);
  for (const key in attr) {
    (tag as any)[key] = attr[key];
  }
  return tag;
}

function translateKey(ev: KeyboardEvent): string {
  switch (ev.key) {
    case 'Alt':
    case 'Control':
    case 'Shift':
    case 'Unidentified':
      return '';
  }
  // Avoid browser tab switch keys:
  if (ev.key >= '0' && ev.key <= '9') return '';

  let name = '';
  if (ev.altKey) name += 'M-';
  if (ev.ctrlKey) name += 'C-';
  if (name.length === 0) return '';
  name += ev.key;
  return name;
}

class ReadLine {
  dom = html('div', { className: 'readline' });
  prompt = html('div', { className: 'prompt' });
  input = html('input', {
    className: 'input',
    spellcheck: false
  }) as HTMLInputElement;
  oncommit = (_: string) => {};

  constructor() {
    this.prompt.innerText = '> ';
    this.dom.appendChild(this.prompt);

    this.dom.appendChild(this.input);

    this.input.onkeydown = ev => {
      this.keydown(ev);
    };
    this.input.onkeypress = ev => {
      this.keypress(ev);
    };
  }

  keydown(ev: KeyboardEvent) {
    const key = translateKey(ev);
    if (!key) return;
    switch (key) {
      case 'C-a':
        this.input.selectionStart = this.input.selectionEnd = 0;
        break;
      case 'C-e':
        const len = this.input.value.length;
        this.input.selectionStart = this.input.selectionEnd = len;
        break;
      case 'C-k':
        this.input.value = this.input.value.substr(
          0,
          this.input.selectionStart!
        );
        break;
      case 'C-J': // browser: inspector
      case 'C-l': // browser: location
      case 'C-r': // browser: reload
        // Allow default handling.
        return;
      default:
        if (key.startsWith('C-Arrow')) return;
        console.log('TODO:', key, ev);
    }
    ev.preventDefault();
  }

  keypress(ev: KeyboardEvent) {
    switch (ev.key) {
      case 'Enter':
        this.oncommit(this.input.value);
        break;
      default:
        return;
    }
    ev.preventDefault();
  }
}

interface Attr {
  fg: number;
  bg: number;
  bright: boolean;
}

/** Decodes a packed attribute number as described in terminal.go. */
function decodeAttr(attr: number): Attr {
  const fg = attr & 0b1111;
  const bg = (attr & 0b11110000) >> 4;
  const bright = (attr & 0x0100) !== 0;
  return { fg, bg, bright };
}

class Term {
  dom = html('pre', { tabIndex: 0 });

  onUpdate(msg: pb.TermText) {
    const children = this.dom.children;
    const row = msg.getRow();
    for (var childCount = children.length; childCount < row + 1; childCount++) {
      this.dom.appendChild(html('div'));
    }
    const child = children[row] as HTMLElement;
    child.innerText = '';
    for (const span of msg.getSpansList()) {
      const { fg, bg, bright } = decodeAttr(span.getAttr());
      const hspan = html('span');
      if (bright) hspan.classList.add(`bright`);
      if (fg > 0) hspan.classList.add(`fg${fg}`);
      if (bg > 0) hspan.classList.add(`bg${bg}`);
      hspan.innerText = span.getText();
      child.appendChild(hspan);
    }
  }
}

class Cell {
  dom = html('div', { className: 'cell' });
  readline = new ReadLine();
  term = new Term();
  onExit = (id: number) => {};

  constructor(public id: number) {
    this.dom.appendChild(this.readline.dom);
    this.dom.appendChild(this.term.dom);

    this.readline.oncommit = cmd => {
      this.readline.input.blur();
      const exec = shell.exec(cmd);
      if (sh.isLocal(exec)) {
        this.term.dom.innerText += exec.output;
        this.onExit(this.id);
      } else {
        spawn(this.id, exec);
      }
    };
  }

  onOutput(msg: pb.Output) {
    if (msg.hasText()) {
      this.term.onUpdate(msg.getText()!);
    }
    if (msg.hasExitCode()) {
      console.log('exit code', msg.getExitCode());
      this.onExit(this.id);
    }
  }
}

class CellStack {
  cells: Cell[] = [];

  addNew() {
    const id = this.cells.length;
    const cell = new Cell(id);
    cell.onExit = (id: number) => {
      this.onExit(id);
    };
    this.cells.push(cell);
    document.body.appendChild(cell.dom);
    cell.readline.input.focus();
  }

  onOutput(msg: pb.Output) {
    this.cells[msg.getCell()].onOutput(msg);
  }

  onExit(id: number) {
    this.addNew();
  }
}

function spawn(id: number, cmd: sh.ExecRemote) {
  if (!ws) return;
  const msg = new pb.RunRequest();
  msg.setCell(id);
  msg.setCwd(cmd.cwd);
  msg.setArgvList(cmd.cmd);
  ws.send(msg.serializeBinary());
}

function handleMessage(ev: MessageEvent) {
  const msg = pb.ServerMsg.deserializeBinary(new Uint8Array(ev.data));
  switch (msg.getMsgCase()) {
    case pb.ServerMsg.MsgCase.OUTPUT: {
      const m = msg.getOutput()!;
      cellStack.onOutput(m);
      break;
    }
  }
}

async function connect(): Promise<WebSocket> {
  const url = new URL('/ws', window.location.href);
  url.protocol = url.protocol.replace('http', 'ws');
  const ws = new WebSocket(url.href);
  ws.binaryType = 'arraybuffer';
  ws.onmessage = handleMessage;
  return new Promise((res, rej) => {
    ws.onopen = () => {
      res(ws);
    };
    ws.onerror = err => {
      rej(err);
    };
  });
}

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

  cellStack = new CellStack();
  cellStack.addNew();

  ws = await connect();
  ws.onclose = ev => {
    console.error(`connection closed: ${ev.code} (${ev.reason})`);
    ws = null;
  };
  ws.onerror = err => {
    console.error(`connection failed: ${err}`);
    ws = null;
  };
}

main().catch(err => {
  console.error(err);
});
