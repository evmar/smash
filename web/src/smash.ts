import * as pb from './smash_pb';
import * as sh from './shell';
import { html, htext } from './html';
import * as readline from './readline';
import { ReadLine } from './readline';
import { Term } from './term';
import { ServerConnection } from './connection';

const shell = new sh.Shell();

interface PendingComplete {
  id: number;
  resolve: (resp: readline.CompleteResponse) => void;
  reject: () => void;
}

class Cell {
  dom = html('div', { className: 'cell' });
  readline = new ReadLine();
  term = new Term();
  running: sh.ExecRemote | null = null;

  delegates = {
    /** Called when the subprocess exits. */
    exit: (id: number, exitCode: number) => {},

    /** Sends a server message. */
    send: (msg: pb.ClientMessage): boolean => false,
  };

  pendingComplete?: PendingComplete;

  constructor(public id: number) {
    this.dom.appendChild(this.readline.dom);
    this.term.delegates = {
      key: (key) => {
        const msg = new pb.ClientMessage();
        key.setCell(this.id);
        msg.setKey(key);
        return this.delegates.send(msg);
      },
    };

    this.readline.oncomplete = async (req) => {
      return new Promise((resolve, reject) => {
        const reqPb = new pb.CompleteRequest();
        reqPb.setId(0);
        reqPb.setCwd(shell.cwd);
        reqPb.setInput(req.input);
        reqPb.setPos(req.pos);
        const msg = new pb.ClientMessage();
        msg.setComplete(reqPb);
        if (!this.delegates.send(msg)) {
          // TOOD
          console.error('send failed');
          reject();
        }
        this.pendingComplete = {
          id: 0,
          resolve,
          reject,
        };
      });
    };

    this.readline.oncommit = (cmd) => {
      const exec = shell.exec(cmd);
      switch (exec.kind) {
        case 'string':
          this.term.dom.innerText += exec.output;
          break;
        case 'table':
          const table = html(
            'table',
            {},
            html(
              'tr',
              {},
              ...exec.headers.map((h) => html('th', {}, htext(h)))
            ),
            ...exec.rows.map((r) =>
              html(
                'tr',
                {},
                ...r.map((t, i) =>
                  html('td', { className: i > 0 ? 'value' : '' }, htext(t))
                )
              )
            )
          );
          this.term.dom = table;
          break;
        case 'remote':
          this.running = exec;
          this.spawn(this.id, exec);
          // The result of spawning will come back in via a message in onOutput().
          break;
      }
      this.dom.appendChild(this.term.dom);
      this.term.dom.focus();
      if (!this.running) {
        this.delegates.exit(this.id, 0);
      }
    };
  }

  spawn(id: number, cmd: sh.ExecRemote): boolean {
    const run = new pb.RunRequest();
    run.setCell(id);
    run.setCwd(cmd.cwd);
    run.setArgvList(cmd.cmd);
    const msg = new pb.ClientMessage();
    msg.setRun(run);
    return this.delegates.send(msg);
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
      this.delegates.exit(this.id, exitCode);
    }
  }

  onCompleteResponse(msg: pb.CompleteResponse) {
    if (!this.pendingComplete) return;
    this.pendingComplete.resolve({
      completions: msg.getCompletionsList(),
      pos: msg.getPos(),
    });
    this.pendingComplete = undefined;
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
  delegates = {
    send: (msg: pb.ClientMessage): boolean => false,
  };

  addNew() {
    const id = this.cells.length;
    const cell = new Cell(id);
    cell.readline.setPrompt(shell.cwdForPrompt());
    cell.delegates = {
      send: this.delegates.send,
      exit: (id: number, exitCode: number) => {
        this.onExit(id, exitCode);
      },
    };
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

  getLastCell(): Cell {
    return this.cells[this.cells.length - 1];
  }

  focus() {
    this.getLastCell().focus();
  }
}

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

  const conn = new ServerConnection();
  const cellStack = new CellStack();

  cellStack.delegates = {
    send: (msg) => conn.send(msg),
  };

  conn.delegates = {
    connect: (hello) => {
      shell.aliases.setAliases(
        new Map<string, string>(hello.getAliasMap().getEntryList())
      );
      shell.env = new Map(hello.getEnvMap().getEntryList());
      shell.init();
    },

    message: (msg) => {
      switch (msg.getMsgCase()) {
        case pb.ServerMsg.MsgCase.COMPLETE: {
          cellStack.getLastCell().onCompleteResponse(msg.getComplete()!);
          break;
        }
        case pb.ServerMsg.MsgCase.OUTPUT: {
          const m = msg.getOutput()!;
          cellStack.onOutput(m);
          break;
        }
        default:
          console.error('unexpected message', msg.toObject());
      }
    },
  };

  await conn.connect();

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
}

main().catch((err) => {
  console.error(err);
});
