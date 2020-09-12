import { CellStack } from './cells';
import { html, htext } from './html';
import * as proto from './proto';
import { Shell } from './shell';

interface Tab {
  /** The tab widget itself, as shown in the tab strip. */
  dom: HTMLElement;

  /** The contents of the tab, shown when the tab is selected. */
  cellStack: CellStack;
}

export class Tabs {
  tabStrip = html('div', { className: 'tabstrip', style: { display: 'none' } });
  dom = html('div', { className: 'tabs' }, this.tabStrip);

  tabs: Tab[] = [];
  sel = -1;
  delegates = {
    send: (msg: proto.ClientMessage) => {},
  };

  addCells(shell: Shell) {
    const tab = this.newTab(shell);
    this.tabs.push(tab);
    this.tabStrip.appendChild(tab.dom);

    if (this.tabs.length > 1) {
      this.tabStrip.style.display = 'flex';
    }

    if (this.sel === -1) {
      this.showTab(0);
    }
  }

  private newTab(shell: Shell): Tab {
    const dom = html('div', { className: 'tab' }, htext('tab'));
    const cellStack = new CellStack(shell);
    cellStack.delegates = {
      send: (msg) => this.delegates.send(msg),
    };
    return { dom, cellStack };
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

  showTab(index: number) {
    if (this.sel === index) return;
    if (this.sel >= 0) {
      this.tabs[this.sel].dom.style.position = 'initial';
      this.dom.removeChild(this.dom.lastChild!);
    }
    this.sel = index;
    this.tabs[index].dom.style.position = 'relative';
    this.dom.appendChild(this.tabs[index].cellStack.dom);
  }

  focus() {
    this.tabs[this.sel].cellStack.focus();
  }
}
