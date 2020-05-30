import * as proto from './proto';
import * as sh from './shell';
import { html, htext } from './html';
import * as readline from './readline';
import { ReadLine } from './readline';
import { Term } from './term';
import { ServerConnection } from './connection';
import { History } from './history';

const shell = new sh.Shell();
const history = new History();

interface PendingComplete {
  id: number;
  resolve: (resp: readline.CompleteResponse) => void;
  reject: () => void;
}

class Cell {
  dom = html('div', { className: 'cell' });
  readline = new ReadLine(history);
  term = new Term();
  running: sh.ExecRemote | null = null;

  delegates = {
    /** Called when the subprocess exits. */
    exit: (id: number, exitCode: number) => {},

    /** Sends a server message. */
    send: (msg: proto.ClientMessage): boolean => false,
  };

  pendingComplete?: PendingComplete;

  constructor(public id: number) {
    this.dom.appendChild(this.readline.dom);
    this.term.delegates = {
      key: (key) => {
        key.cell = this.id;
        const msg: proto.ClientMessage = new proto.ClientMessage(key);
        return this.delegates.send(msg);
      },
    };

    this.readline.delegates = {
      oncomplete: async (req) => {
        return new Promise((resolve, reject) => {
          const reqPb = new proto.CompleteRequest({
            id: 0,
            cwd: shell.cwd,
            input: req.input,
            pos: req.pos,
          });
          const msg = new proto.ClientMessage(reqPb);
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
      },

      oncommit: (cmd) => {
        const exec = shell.exec(cmd);
        switch (exec.kind) {
          case 'string':
            this.term.dom.innerText += exec.output;
            break;
          case 'table':
            this.term.dom = this.renderTable(exec);
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
      },
    };
  }

  private renderTable(exec: sh.TableOutput) {
    return html(
      'table',
      {},
      html('tr', {}, ...exec.headers.map((h) => html('th', {}, htext(h)))),
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
  }

  spawn(id: number, cmd: sh.ExecRemote): boolean {
    const run = new proto.RunRequest({
      cell: id,
      cwd: cmd.cwd,
      argv: cmd.cmd,
    });
    const msg = new proto.ClientMessage(run);
    return this.delegates.send(msg);
  }

  onOutput(msg: proto.Output) {
    if (msg.alt instanceof proto.CmdError) {
      // error
      this.term.showError(msg.alt.error);
    } else if (msg.alt instanceof proto.TermUpdate) {
      this.term.onUpdate(msg.alt);
    } else {
      // exit code
      // Command completed.
      const exitCode = msg.alt.exitCode;
      if (this.running && this.running.onComplete) {
        this.running.onComplete(exitCode);
      }
      this.running = null;
      this.term.showCursor(false);
      this.term.preventFocus();
      this.delegates.exit(this.id, exitCode);
    }

    // If the terminal was in focus, scroll to the bottom.
    // TODO: handle the case where the user has scrolled back.
    if (document.activeElement === this.term.dom) {
      document.scrollingElement!.scrollIntoView({
        block: 'end',
      });
    }
  }

  onCompleteResponse(msg: proto.CompleteResponse) {
    if (!this.pendingComplete) return;
    this.pendingComplete.resolve({
      completions: msg.completions,
      pos: msg.pos,
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
    send: (msg: proto.ClientMessage): boolean => false,
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

  onOutput(msg: proto.CellOutput) {
    this.cells[msg.cell].onOutput(msg.output);
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
        new Map<string, string>(hello.alias.map(({ key, val }) => [key, val]))
      );
      shell.env = new Map(hello.env.map(({ key, val }) => [key, val]));
      shell.init();
    },

    message: (msg) => {
      if (msg.alt instanceof proto.CompleteResponse) {
        cellStack.getLastCell().onCompleteResponse(msg.alt);
      } else if (msg.alt instanceof proto.CellOutput) {
        cellStack.onOutput(msg.alt);
      } else {
        console.error('unexpected message', msg);
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
