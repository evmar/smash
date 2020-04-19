import { html, htext } from './html';
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
  Tab: '\x09',
  Enter: '\x0d',
  Escape: '\x1b',

  // Add these keys to the map because we warn on any key
  // not in the map.
  Alt: '',
  Control: '',
  Shift: '',
};

/**
 * Client side DOM of terminal emulation.
 * 
 * The actual vt100 etc. emulation happens on the server.
 * This client receives screen updates and forwards keystrokes.
 */
export class Term {
  dom = html('pre', { tabIndex: 0, className: 'term' });
  cursor = html('div', { className: 'term-cursor' });
  cellSize = { width: 0, height: 0 };

  delegates = {
    /** Sends a keyboard event to the terminal's subprocess. */
    key: (msg: pb.KeyEvent): boolean => false,
  }

  constructor() {
    this.dom.onkeydown = (e) => this.onKeyDown(e);
    this.dom.onkeypress = (e) => this.onKeyPress(e);
    this.dom.appendChild(this.cursor);
    this.measure();
  }

  measure() {
    document.body.appendChild(this.dom);
    this.cursor.innerText = 'A';
    const { width, height } = getComputedStyle(this.cursor);
    document.body.removeChild(this.dom);
    this.cursor.innerText = '';
    this.cursor.style.width = width;
    this.cursor.style.height = height;
    this.cellSize.width = Number(width!.replace('px', ''));
    this.cellSize.height = Number(height!.replace('px', ''));
  }

  focus() {
    this.dom.focus();
  }

  onUpdate(msg: pb.TermUpdate) {
    const children = this.dom.children;
    for (const rowSpans of msg.getRowsList()) {
      const row = rowSpans.getRow() + 1; // +1 to avoid this.cursor
      for (
        var childCount = children.length;
        childCount < row + 1;
        childCount++
      ) {
        this.dom.appendChild(html('div', {}, htext(' ')));
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
    const cursor = msg.getCursor();
    if (cursor) {
      this.showCursor(!cursor.getHidden());
      this.cursor.style.left = cursor.getCol() * this.cellSize.width + 'px';
      this.cursor.style.top = cursor.getRow() * this.cellSize.height + 'px';
    }
  }

  showCursor(show: boolean) {
    this.cursor.style.display = show ? 'block' : 'none';
  }

  showError(msg: string) {
    const div = html('div');
    div.innerText = msg;
    this.dom.appendChild(div);
  }

  sendKeys(keys: string) {
    const msg = new pb.KeyEvent();
    msg.setKeys(keys);
    return this.delegates.key(msg);
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
      if (send === undefined) console.log('term: unknown key:', key);
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
