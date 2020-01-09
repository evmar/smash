import * as pb from './smash_pb';
import * as sh from './shell';
import { html, htext } from './html';
import { ReadLine } from './readline';
import { Term } from './term';
import * as jspb from 'google-protobuf';

class ServerConnection {
  ws: WebSocket | null = null;
  onMessage = (msg: pb.ServerMsg) => {};
  errorDom: HTMLElement | null = null;

  async connect(): Promise<void> {
    const url = new URL('/ws', window.location.href);
    url.protocol = url.protocol.replace('http', 'ws');
    const ws = new WebSocket(url.href);
    ws.binaryType = 'arraybuffer';
    ws.onmessage = event => {
      const msg = pb.ServerMsg.deserializeBinary(new Uint8Array(event.data));
      this.onMessage(msg);
    };
    this.ws = await new Promise(res => {
      ws.onopen = () => {
        res(ws);
      };
      ws.onerror = (err: Event) => {
        this.showError(`websocket connection failed`);
        res(null);
      };
    });

    if (!this.ws) return;

    this.ws.onopen = ev => {
      console.error(`unexpected ws open:`, ev);
    };
    this.ws.onclose = ev => {
      let msg = 'connection closed';
      if (ev.reason) msg += `: ${ev.reason}`;
      this.showError(msg);
      this.ws = null;
    };
    this.ws.onerror = err => {
      this.showError(`connection error: ${err}`);
      this.ws = null;
    };
  }

  reconnect() {
    if (!this.errorDom) return;
    document.body.removeChild(this.errorDom);
    this.errorDom = null;
    this.connect();
  }

  send(msg: pb.ClientMessage): boolean {
    if (!this.ws) return false;
    this.ws.send(msg.serializeBinary());
    return true;
  }

  spawn(id: number, cmd: sh.ExecRemote): boolean {
    const run = new pb.RunRequest();
    run.setCell(id);
    run.setCwd(cmd.cwd);
    run.setArgvList(cmd.cmd);
    const msg = new pb.ClientMessage();
    msg.setRun(run);
    return this.send(msg);
  }

  showError(msg: string) {
    console.error(msg);
    if (!this.errorDom) {
      this.errorDom = html(
        'div',
        { className: 'error-popup' },
        html('div', {}, document.createTextNode(msg)),
        html('div', { style: { width: '1ex' } }),
        html(
          'button',
          {
            onclick: () => {
              this.reconnect();
            }
          },
          document.createTextNode('reconnect')
        )
      );
      document.body.appendChild(this.errorDom);
    }
  }
}

const conn = new ServerConnection();
const shell = new sh.Shell();

class Cell {
  dom = html('div', { className: 'cell' });
  readline = new ReadLine();
  term = new Term();
  running: sh.ExecRemote | null = null;
  onExit = (id: number, exitCode: number) => {};
  send = (msg: pb.ClientMessage): boolean => false;

  constructor(public id: number) {
    this.dom.appendChild(this.readline.dom);
    this.term.send = key => {
      const msg = new pb.ClientMessage();
      key.setCell(this.id);
      msg.setKey(key);
      return this.send(msg);
    };

    this.readline.oncommit = cmd => {
      const exec = shell.exec(cmd);
      switch (exec.kind) {
        case 'string':
          this.term.dom.innerText += exec.output;
          this.onExit(this.id, 0);
          break;
        case 'table':
          const table = html(
            'table',
            {},
            html('tr', {}, ...exec.headers.map(h => html('th', {}, htext(h)))),
            ...exec.rows.map(r =>
              html('tr', {}, ...r.map(t => html('td', {}, htext(t))))
            )
          );
          this.term.dom = table;
          this.onExit(this.id, 0);
          break;
        case 'remote':
          this.running = exec;
          conn.spawn(this.id, exec);
        // The result of spawning will come back in via a message in onOutput().
      }
      this.dom.appendChild(this.term.dom);
      this.term.dom.focus();
    };
  }

  onOutput(msg: pb.Output) {
    if (msg.hasTermUpdate()) {
      this.term.onUpdate(msg.getTermUpdate()!);
    }
    if (msg.hasError()) {
      this.term.showError(msg.getError());
    }
    if (msg.hasExitCode()) {
      // Command completed.
      const exitCode = msg.getExitCode();
      if (this.running && this.running.onComplete) {
        this.running.onComplete(exitCode);
      }
      this.running = null;
      this.term.showCursor(false);
      this.onExit(this.id, exitCode);
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
  send = (msg: pb.ClientMessage): boolean => false;

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

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

  async function connect(): Promise<pb.Hello> {
    return new Promise(async (resolve, reject) => {
      conn.onMessage = msg => {
        if (msg.getMsgCase() !== pb.ServerMsg.MsgCase.HELLO) {
          reject(`unexpected message ${msg.toObject()}`);
          return;
        }
        resolve(msg.getHello());
      };
      await conn.connect();
    });
  }
  const hello = await connect();
  shell.aliases.setAliases(
    new Map<string, string>(hello.getAliasMap().getEntryList())
  );

  const cellStack = new CellStack();

  conn.onMessage = msg => {
    switch (msg.getMsgCase()) {
      case pb.ServerMsg.MsgCase.OUTPUT: {
        const m = msg.getOutput()!;
        cellStack.onOutput(m);
        break;
      }
      default:
        console.error('unexpected message', msg.toObject());
    }
  };

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

  cellStack.send = msg => conn.send(msg);
}

main().catch(err => {
  console.error(err);
});
