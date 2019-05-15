import * as pb from './smash_pb';
import * as sh from './shell';
import { html } from './html';
import { ReadLine } from './readline';
import { Term } from './term';

let ws: WebSocket | null = null;
let shell = new sh.Shell();
let cellStack: CellStack;

class Cell {
  dom = html('div', { className: 'cell' });
  readline = new ReadLine();
  term = new Term();
  running = false;
  onExit = (id: number, exitCode: number) => {};
  send = (msg: pb.ClientMessage) => {};

  constructor(public id: number) {
    this.dom.appendChild(this.readline.dom);
    this.term.send = key => {
      const msg = new pb.ClientMessage();
      key.setCell(this.id);
      msg.setKey(key);
      this.send(msg);
    };

    this.readline.oncommit = cmd => {
      this.dom.appendChild(this.term.dom);
      this.term.dom.focus();
      const exec = shell.exec(cmd);
      if (sh.isLocal(exec)) {
        this.term.dom.innerText += exec.output;
        this.onExit(this.id, 0);
      } else {
        this.running = true;
        spawn(this.id, exec);
      }
    };
  }

  onOutput(msg: pb.Output) {
    if (msg.hasTermUpdate()) {
      this.term.onUpdate(msg.getTermUpdate()!);
    }
    if (msg.hasExitCode()) {
      this.running = false;
      this.term.showCursor(false);
      this.onExit(this.id, msg.getExitCode());
    }
  }

  focus() {
    if (this.running) {
      this.term.focus();
    } else {
      this.readline.focus();
    }
  }
}

class CellStack {
  cells: Cell[] = [];
  send = (msg: pb.ClientMessage) => {};

  addNew() {
    const id = this.cells.length;
    const cell = new Cell(id);
    cell.onExit = (id: number, exitCode: number) => {
      this.onExit(id, exitCode);
    };
    cell.send = msg => this.send(msg);
    this.cells.push(cell);
    document.body.appendChild(cell.dom);
    cell.readline.input.focus();
  }

  onOutput(msg: pb.Output) {
    this.cells[msg.getCell()].onOutput(msg);
  }

  onExit(id: number, exitCode: number) {
    this.addNew();
  }

  focus() {
    this.cells[this.cells.length - 1].focus();
  }
}

function spawn(id: number, cmd: sh.ExecRemote) {
  if (!ws) return;
  const run = new pb.RunRequest();
  run.setCell(id);
  run.setCwd(cmd.cwd);
  run.setArgvList(cmd.cmd);
  const msg = new pb.ClientMessage();
  msg.setRun(run);
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

  // Clicking on the page, if it tries to focus the document body,
  // should redirect focus to the relevant place in the cell stack.
  // This approach feels hacky but I experimented with focus events
  // and couldn't get the desired behavior.
  document.addEventListener('click', () => {
    if (document.activeElement === document.body) {
      cellStack.focus();
    }
  });

  ws = await connect();
  ws.onclose = ev => {
    console.error(`connection closed: ${ev.code} (${ev.reason})`);
    ws = null;
  };
  ws.onerror = err => {
    console.error(`connection failed: ${err}`);
    ws = null;
  };
  cellStack.send = msg => {
    if (!ws) return;
    ws.send(msg.serializeBinary());
  };
}

main().catch(err => {
  console.error(err);
});
