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
  /** Did the subprocess produce any output? */
  didOutput = false;
  running: sh.ExecRemote | null = null;

  delegates = {
    /** Called when the subprocess exits. */
    exit: (id: number, exitCode: number) => {},

    /** Sends a server message. */
    send: (msg: proto.ClientMessage) => {},
  };

  pendingComplete?: PendingComplete;

  constructor(readonly id: number, readonly shell: Shell) {
    this.dom.appendChild(this.readline.dom);
    this.term.delegates = {
      key: (key) => {
        key.cell = this.id;
        this.delegates.send({ tag: 'KeyEvent', val: key });
      },
    };

    this.readline.delegates = {
      oncomplete: async (req) => {
        return new Promise((resolve, reject) => {
          const reqProto: proto.CompleteRequest = {
            id: 0,
            cwd: shell.cwd,
            input: req.input,
            pos: req.pos,
          };
          const msg: proto.ClientMessage = {
            tag: 'CompleteRequest',
            val: reqProto,
          };
          this.delegates.send(msg);
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
            this.term.dom.innerText = exec.output;
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

  spawn(id: number, cmd: sh.ExecRemote) {
    const run: proto.RunRequest = {
      cell: id,
      cwd: cmd.cwd,
      argv: cmd.cmd,
    };
    this.delegates.send({ tag: 'RunRequest', val: run });
  }

  onOutput(msg: proto.Output) {
    switch (msg.tag) {
      case 'CmdError':
        // error; exit code will come later.
        this.dom.appendChild(html('div', {}, htext(msg.val.error)));
        break;
      case 'TermUpdate':
        this.didOutput = true;
        this.term.onUpdate(msg.val);
        break;
      case 'Exit':
        // exit code
        // Command completed.
        const exitCode = msg.val.exitCode;
        if (this.running && this.running.onComplete) {
          this.running.onComplete(exitCode);
        }
        this.running = null;
        this.term.showCursor(false);
        this.term.preventFocus();
        if (!this.didOutput) {
          // Remove the vertical space of the terminal.
          this.term.dom.innerText = '';
        }
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

function scrollToBottom(el: HTMLElement) {
  el.scrollIntoView({
    block: 'end',
  });
}

export class CellStack {
  dom = html('div', { className: 'cellstack' });
  cells: Cell[] = [];
  delegates = {
    send: (msg: proto.ClientMessage) => {},
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
    scrollToBottom(cell.dom);
  }

  onOutput(msg: proto.CellOutput) {
    const cell = this.cells[msg.cell];
    cell.onOutput(msg.output);
    if (msg.cell === this.cells.length - 1) {
      scrollToBottom(cell.dom);
    }
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
