import { html, htext } from './html';

export function translateKey(ev: KeyboardEvent): string {
  switch (ev.key) {
    case 'Alt':
    case 'Control':
    case 'Shift':
    case 'Unidentified':
      return '';
  }
  // Avoid browser tab switch keys:
  if (ev.key >= '0' && ev.key <= '9') return '';

  let name = '';
  if (ev.altKey) name += 'M-';
  if (ev.ctrlKey) name += 'C-';
  if (ev.shiftKey && ev.key.length > 1) name += 'S-';
  name += ev.key;
  return name;
}

export interface CompleteRequest {
  input: string;
  pos: number;
}

export interface CompleteResponse {
  completions: string[];
  pos: number;
}

class CompletePopup {
  dom = html('div', { className: 'popup', style: { overflow: 'hidden' } });
  textSize!: { width: number; height: number };
  selection = -1;

  delegates = {
    oncommit: (text: string, pos: number): void => {},
  };

  constructor(readonly req: CompleteRequest, readonly resp: CompleteResponse) {}

  show(parent: HTMLElement) {
    this.textSize = this.measure(
      parent,
      this.req.input.substring(0, this.resp.pos) + '\u200b'
    );

    for (const comp of this.resp.completions) {
      const dom = html('div', { className: 'completion' }, htext(comp));
      // Listen to mousedown because if we listen to click, the click causes
      // the input field to lose focus.
      dom.addEventListener('mousedown', (event) => {
        this.delegates.oncommit(comp, this.resp.pos);
        event.preventDefault();
      });
      this.dom.appendChild(dom);
    }
    parent.appendChild(this.dom);
    this.position();
    this.selectCompletion(0);
    this.dom.focus();
  }

  /** Measures the size of the given text as if it were contained in the parent. */
  private measure(
    parent: HTMLElement,
    text: string
  ): { width: number; height: number } {
    const measure = html(
      'div',
      {
        style: {
          position: 'absolute',
          visibility: 'hidden',
          whiteSpace: 'pre',
        },
      },
      htext(text)
    );
    parent.appendChild(measure);
    const { width, height } = getComputedStyle(measure);
    parent.removeChild(measure);
    return { width: parseFloat(width), height: parseFloat(height) };
  }

  /** Positions this.dom. */
  private position() {
    // Careful about units here.  The element is positioned relative to the input
    // box, but we want to measure things in terms of whether they fit in the current
    // viewport.
    //
    // Also, the popup may not fit.  Options in order of preference:
    // 1. Pop up below, if it fits.
    // 2. Pop up above, if it fits.
    // 3. Pop up in whichever side has more space, but truncated.

    // promptX/promptY are in viewport coordinates.
    const promptY = (this.dom.parentNode as HTMLElement).getClientRects()[0].y;
    const popupHeight = this.dom.offsetHeight;

    const spaceAbove = promptY;
    const spaceBelow = window.innerHeight - (promptY + this.textSize.height);

    let placeBelow: boolean;
    if (spaceBelow >= popupHeight) {
      placeBelow = true;
    } else if (spaceAbove >= popupHeight) {
      placeBelow = false;
    } else {
      placeBelow = spaceBelow >= spaceAbove;
    }

    const popupPaddingY = 2 + 2; // 2 above, 2 below
    const popupShadowY = 4; // arbitrary fudge factor
    const popupSizeMargin = popupPaddingY + popupShadowY;

    if (placeBelow) {
      this.dom.style.top = `${this.textSize.height}px`;
      this.dom.style.bottom = '';
      this.dom.style.height =
        spaceBelow >= popupHeight ? '' : `${spaceBelow - popupSizeMargin}px`;
    } else {
      this.dom.style.top = '';
      this.dom.style.bottom = `${this.textSize.height}px`;
      this.dom.style.height =
        spaceAbove >= popupHeight ? '' : `${spaceAbove - popupSizeMargin}px`;
    }

    const popupPaddingLeft = 4;
    this.dom.style.left = `${this.textSize.width - popupPaddingLeft}px`;
  }

  hide() {
    this.dom.parentNode!.removeChild(this.dom);
  }

  private selectCompletion(index: number) {
    if (this.selection !== -1) {
      this.dom.children[this.selection].classList.remove('selected');
    }
    this.selection =
      (index + this.resp.completions.length) % this.resp.completions.length;
    this.dom.children[this.selection].classList.add('selected');
  }

  /** @param key The key name as produced by translateKey(). */
  handleKey(key: string): boolean {
    switch (key) {
      case 'ArrowDown':
      case 'Tab':
      case 'C-n':
        this.selectCompletion(this.selection + 1);
        return true;
      case 'ArrowUp':
      case 'S-Tab':
      case 'C-p':
        this.selectCompletion(this.selection - 1);
        return true;
      case 'Enter':
        this.delegates.oncommit(
          this.resp.completions[this.selection],
          this.resp.pos
        );
        return true;
      case 'Escape':
        this.delegates.oncommit('', this.resp.pos);
        return true;
    }
    return false; // Pop down on any other key.
  }
}

/** Returns the length of the longest prefix shared by all input strings. */
function longestSharedPrefixLength(strs: string[]): number {
  for (let len = 0; ; len++) {
    let c = -1;
    for (const str of strs) {
      if (len === str.length) return len;
      if (c === -1) c = str.charCodeAt(len);
      else if (str.charCodeAt(len) !== c) return len;
    }
  }
}

export function backwardWordBoundary(text: string, pos: number): number {
  // If at a word start already, skip preceding whitespace.
  for (; pos > 0; pos--) {
    if (text.charAt(pos - 1) !== ' ') break;
  }
  // Skip to the beginning of the current word.
  for (; pos > 0; pos--) {
    if (text.charAt(pos - 1) === ' ') break;
  }
  return pos;
}

function forwardWordBoundary(text: string, pos: number): number {
  for (; pos < text.length; pos++) {
    if (text.charAt(pos) === ' ') break;
  }
  for (; pos < text.length; pos++) {
    if (text.charAt(pos) !== ' ') break;
  }
  return pos;
}

export interface InputState {
  text: string;
  start: number;
  end: number;
}

export interface InputHandler {
  onEnter(state: InputState): void;
  tabComplete(state: InputState): void;
  setText(text: string): void;
  setPos(pos: number): void;
  showHistory(delta: -1 | 0 | 1): void;
}

export function interpretKey(
  state: InputState,
  key: string,
  handler: InputHandler
): boolean {
  const { text, start } = state;
  switch (key) {
    case 'Enter':
      handler.onEnter(state);
      return true;
    case 'Tab':
      handler.tabComplete(state);
      return true;
    case 'Delete': // At least on ChromeOS, this is M-Backspace.
    case 'M-Backspace': {
      // backward-kill-word
      const wordStart = backwardWordBoundary(text, start);
      handler.setPos(wordStart);
      handler.setText(text.substring(0, wordStart) + text.substring(start));
      return true;
    }
    case 'C-a':
    case 'Home':
      handler.setPos(0);
      return true;
    case 'C-b':
      handler.setPos(start - 1);
      return true;
    case 'M-b':
      handler.setPos(backwardWordBoundary(text, start));
      return true;
    case 'M-d': {
      const delEnd = forwardWordBoundary(text, start);
      handler.setText(text.substring(0, start) + text.substring(delEnd));
      return true;
    }
    case 'C-e':
    case 'End':
      handler.setPos(text.length);
      return true;
    case 'C-f':
      handler.setPos(start + 1);
      return true;
    case 'M-f':
      handler.setPos(forwardWordBoundary(text, start));
      return true;
    case 'C-k':
      handler.setText(text.substr(0, start));
      return true;
    case 'C-n':
    case 'ArrowDown':
      handler.showHistory(-1);
      return true;
    case 'C-p':
    case 'ArrowUp':
      handler.showHistory(1);
      return true;
    case 'C-u':
      handler.setText(text.substr(start));
      return true;

    case 'C-x': // browser: cut
    case 'C-c': // browser: copy
    case 'C-v': // browser: paste
    case 'C-J': // browser: inspector
    case 'C-l': // browser: location
    case 'C-R': // browser: reload
      // Allow default handling.
      return false;
    default:
      handler.showHistory(0);
      return false;
  }
}

export interface History {
  add(cmd: string): void;
  get(ofs: number): string | undefined;
}

export class ReadLine {
  dom = html('div', { className: 'readline' });
  prompt = html('div', { className: 'prompt' });
  inputBox = html('div', { className: 'input-box' });
  input = html('input', {
    spellcheck: false,
  }) as HTMLInputElement;

  delegates = {
    oncommit: (text: string): void => {},
    oncomplete: async (req: CompleteRequest): Promise<CompleteResponse> => {
      throw 'notimpl';
    },
  };

  pendingComplete: Promise<CompleteResponse> | undefined;
  popup: CompletePopup | undefined;

  /** Offset into the history: "we have gone N commands back". */
  historyPosition = 0;

  /**
   * The selection span at time of last blur.
   * This is restored on focus, to defeat the browser behavior of
   * select all on focus.
   */
  selection: [number, number] = [0, 0];

  constructor(private history: History) {
    this.dom.appendChild(this.prompt);

    this.inputBox.appendChild(this.input);
    this.dom.appendChild(this.inputBox);

    this.input.onkeydown = (ev) => {
      const key = translateKey(ev);
      if (!key) return;
      if (this.handleKey(key)) ev.preventDefault();
    };
    this.input.onkeypress = (ev) => {
      const key = ev.key;
      if (!key) return;
      if (this.handleKey(key)) ev.preventDefault();
    };

    // Catch focus/blur events, per docs on this.selection.
    this.input.addEventListener('blur', () => {
      this.selection = [this.input.selectionStart!, this.input.selectionEnd!];
      this.pendingComplete = undefined;
      this.hidePopup();
    });
    this.input.addEventListener('focus', () => {
      [this.input.selectionStart, this.input.selectionEnd] = this.selection;
    });
  }

  setPrompt(text: string) {
    this.prompt.innerText = `${text}$ `;
  }

  setText(text: string) {
    this.input.value = text;
    this.historyPosition = 0;
  }

  setPos(pos: number) {
    pos = Math.max(0, Math.min(this.input.value.length, pos));
    this.input.selectionStart = this.input.selectionEnd = pos;
  }

  showHistory(delta: -1 | 0 | 1) {
    switch (delta) {
      case -1: {
        if (this.historyPosition === 0) return;
        this.historyPosition--;
        const cmd = this.history.get(this.historyPosition) || '';
        this.input.value = cmd;
        return;
      }
      case 1: {
        const cmd = this.history.get(this.historyPosition + 1);
        if (!cmd) return;
        this.historyPosition++;
        this.input.value = cmd;
        return;
      }
      case 0:
        this.historyPosition = 0;
        return;
    }
  }

  focus() {
    this.input.focus();
  }

  hidePopup() {
    if (!this.popup) return;
    this.popup.hide();
    this.popup = undefined;
  }

  /** @param key The key name as produced by translateKey(). */
  handleKey(key: string): boolean {
    if (this.popup && this.popup.handleKey(key)) return true;
    if (this.pendingComplete) this.pendingComplete = undefined;
    this.hidePopup();

    const state: InputState = {
      text: this.input.value,
      start: this.input.selectionStart ?? 0,
      end: this.input.selectionEnd ?? 0,
    };
    return interpretKey(state, key, this);
  }

  tabComplete(state: InputState) {
    const pos = state.start;
    const req: CompleteRequest = { input: state.text, pos };
    const pending = (this.pendingComplete = this.delegates.oncomplete(req));
    pending.then((resp) => {
      if (pending !== this.pendingComplete) return;
      this.pendingComplete = undefined;
      if (resp.completions.length === 0) return;
      const len = longestSharedPrefixLength(resp.completions);
      if (len > 0) {
        this.applyCompletion(resp.completions[0].substring(0, len), resp.pos);
      }
      // If there was only one completion, it's already been applied, so
      // there is nothing else to do.
      if (resp.completions.length > 1) {
        // Show a popup for the completions.
        this.popup = new CompletePopup(req, resp);
        this.popup.show(this.inputBox);
        this.popup.delegates = {
          oncommit: (text: string, pos: number) => {
            this.applyCompletion(text, pos);
            this.hidePopup();
          },
        };
      }
    });
  }

  applyCompletion(text: string, pos: number) {
    // The completion for a partial input may include some of that
    // partial input.  Elide any text from the completion that already
    // exists in the input at that same position.
    let overlap = 0;
    while (
      pos + overlap < this.input.value.length &&
      this.input.value[pos + overlap] === text[overlap]
    ) {
      overlap++;
    }
    this.setText(
      this.input.value.substring(0, pos) +
        text +
        this.input.value.substring(pos + overlap)
    );
  }

  onEnter() {
    const text = this.input.value;
    this.history.add(text);
    this.delegates.oncommit(text);
  }
}
