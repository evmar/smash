import { CellStack } from './cells';
import * as proto from './proto';
import { html } from './html';
import { Shell } from './shell';

interface Tab {
  cellStack: CellStack;
}

export class Tabs {
  dom = html('div', { className: 'tabs' });
  tabs: Tab[] = [];
  delegates = {
    send: (msg: proto.ClientMessage) => {
      return false;
    },
  };

  addCells(shell: Shell) {
    if (this.tabs.length > 0) throw new Error('multiple newclls');
    const tab = this.newCells(shell);
    this.tabs.push(tab);
    this.dom.appendChild(tab.cellStack.dom);
  }

  private newCells(shell: Shell): Tab {
    const cellStack = new CellStack(shell);
    cellStack.delegates = {
      send: (msg) => this.delegates.send(msg),
    };
    return { cellStack };
  }

  handleMessage(msg: proto.ServerMsg): boolean {
    const cellStack = this.tabs[0].cellStack;
    if (msg.alt instanceof proto.CompleteResponse) {
      cellStack.getLastCell().onCompleteResponse(msg.alt);
      return true;
    } else if (msg.alt instanceof proto.CellOutput) {
      cellStack.onOutput(msg.alt);
      return true;
    }
    return false;
  }

  focus() {
    this.tabs[0].cellStack.focus();
  }
}
