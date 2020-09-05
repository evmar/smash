import { CellStack } from './cells';
import { html } from './html';
import * as proto from './proto';
import { Shell } from './shell';

interface Tab {
  cellStack: CellStack;
}

export class Tabs {
  dom = html('div', { className: 'tabs' });
  tabs: Tab[] = [];
  delegates = {
    send: (msg: proto.ClientMessage) => {},
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
    switch (msg.tag) {
      case 'CompleteResponse':
        cellStack.getLastCell().onCompleteResponse(msg.val);
        return true;
      case 'CellOutput':
        cellStack.onOutput(msg.val);
        return true;
    }
    return false;
  }

  focus() {
    this.tabs[0].cellStack.focus();
  }
}
