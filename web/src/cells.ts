import { History } from './history';
import { htext, html } from './html';
import * as proto from './proto';
import * as readline from './readline';
import { ReadLine } from './readline';
import * as sh from './shell';
import { Shell } from './shell';
import { Term } from './term';

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

  constructor(readonly id: number, readonly shell: Shell) {
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

function scrollToBottom() {
  document.scrollingElement!.scrollIntoView({
    block: 'end',
  });
}

export class CellStack {
  dom = html('div', { className: 'cellstack' });
  cells: Cell[] = [];
  delegates = {
    send: (msg: proto.ClientMessage): boolean => false,
  };

  constructor(readonly shell: Shell) {
    this.addNew();
  }

  addNew() {
    const id = this.cells.length;
    const cell = new Cell(id, this.shell);
    cell.readline.setPrompt(this.shell.cwdForPrompt());
    cell.delegates = {
      send: (msg) => this.delegates.send(msg),
      exit: (id: number, exitCode: number) => {
        this.onExit(id, exitCode);
      },
    };
    this.cells.push(cell);
    this.dom.appendChild(cell.dom);
    cell.readline.input.focus();
    scrollToBottom();
  }

  onOutput(msg: proto.CellOutput) {
    this.cells[msg.cell].onOutput(msg.output);
    scrollToBottom();
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
