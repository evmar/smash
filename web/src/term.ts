import { html, htext } from './html';
import * as proto from './proto';
import { translateKey } from './readline';

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

  'C-c': '\x03',
  'C-d': '\x04',

  Backspace: '\x08',
  Tab: '\x09',
  Enter: '\x0d',
  'C-[': '\x1b',
  Escape: '\x1b',
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
    key: (msg: proto.KeyEvent) => {},
  };

  constructor() {
    this.dom.onkeydown = (e) => this.onKeyDown(e);
    this.dom.onkeypress = (e) => this.onKeyPress(e);
    this.dom.appendChild(this.cursor);
    this.measure();
    // Create initial empty line, for height.
    // This will be replaced as soon as an update comes in.
    this.dom.appendChild(html('div', {}, htext(' ')));
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

  preventFocus() {
    this.dom.removeAttribute('tabindex');
  }

  onUpdate(msg: proto.TermUpdate) {
    let childIdx = 0;
    let child = this.dom.children[1] as HTMLElement; // avoid this.cursor
    for (const rowSpans of msg.rows) {
      const row = rowSpans.row;
      for (; childIdx < row; childIdx++) {
        if (!child.nextSibling) {
          this.dom.appendChild(html('div', {}, htext(' ')));
        }
        child = child.nextSibling! as HTMLElement;
      }
      const spans = rowSpans.spans;
      if (spans.length === 0) {
        // Empty line. Set text to something non-empty so the div isn't
        // collapsed.
        child.innerText = ' ';
      } else {
        child.innerText = '';
        for (const span of spans) {
          const { fg, bg, bright } = decodeAttr(span.attr);
          const hspan = html('span');
          if (bright) hspan.classList.add(`bright`);
          if (fg > 0) hspan.classList.add(`fg${fg}`);
          if (bg > 0) hspan.classList.add(`bg${bg}`);
          hspan.innerText = span.text;
          child.appendChild(hspan);
        }
      }
    }
    const cursor = msg.cursor;
    if (cursor) {
      this.showCursor(!cursor.hidden);
      this.cursor.style.left = cursor.col * this.cellSize.width + 'px';
      this.cursor.style.top = cursor.row * this.cellSize.height + 'px';
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
    const msg = new proto.KeyEvent({ cell: 0, keys });
    return this.delegates.key(msg);
  }

  onKeyDown(ev: KeyboardEvent) {
    let key = translateKey(ev);
    const send = termKeyMap[key];
    if (!send) return;
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
