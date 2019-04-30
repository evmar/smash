import { html } from './html';
import * as pb from './smash_pb';

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

const termKeyMap: { [key: string]: string } = {
  ArrowUp: '\x1b[A',
  ArrowDown: '\x1b[B',
  ArrowRight: '\x1b[C',
  ArrowLeft: '\x1b[D',

  Backspace: '\x08',
  Enter: '\x0d',
  Escape: '\x1b',

  // Add these keys to the map because we warn on any key
  // not in the map.
  Alt: '',
  Control: '',
  Shift: ''
};

export class Term {
  dom = html('pre', { tabIndex: 0 });
  send = (msg: pb.KeyEvent) => {};

  constructor() {
    this.dom.onkeydown = e => this.onKeyDown(e);
    this.dom.onkeypress = e => this.onKeyPress(e);
  }

  onUpdate(msg: pb.TermText) {
    const children = this.dom.children;
    for (const rowSpans of msg.getRowsList()) {
      const row = rowSpans.getRow();
      for (
        var childCount = children.length;
        childCount < row + 1;
        childCount++
      ) {
        this.dom.appendChild(html('div'));
      }
      const child = children[row] as HTMLElement;
      const spans = rowSpans.getSpansList();
      if (spans.length === 0) {
        // Empty line. Set text to something non-empty so the div isn't
        // collapsed.
        child.innerText = ' ';
      } else {
        child.innerText = '';
        for (const span of spans) {
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
  }

  sendKeys(keys: string) {
    const msg = new pb.KeyEvent();
    msg.setKeys(keys);
    this.send(msg);
  }

  onKeyDown(ev: KeyboardEvent) {
    let key = ev.key;
    switch (key) {
      case 'BracketLeft':
        if (ev.ctrlKey) key = 'Escape';
        break;
    }

    if (key.length === 1) return;

    const send = termKeyMap[key];
    if (!send) {
      if (send === undefined) console.log('unknown key:', key);
      return;
    }
    this.sendKeys(send);
    ev.preventDefault();
  }

  onKeyPress(ev: KeyboardEvent) {
    if (ev.key.length !== 1) {
      console.log('long press', ev.key);
      return;
    }
    this.sendKeys(ev.key);
    ev.preventDefault();
  }
}
