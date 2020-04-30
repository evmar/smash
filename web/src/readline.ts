import { html, htext } from './html';

function translateKey(ev: KeyboardEvent): string {
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

  delegates = {
    oncommit: (text: string, pos: number): void => {},
  };

  constructor(readonly req: CompleteRequest, readonly resp: CompleteResponse) {}

  show(parent: HTMLElement) {
    this.textSize = this.measure(
      parent,
      this.req.input.substring(0, this.resp.pos) + '\u200b'
    );

    this.dom.innerText = this.resp.completions.join('\n');
    parent.appendChild(this.dom);
    this.position();
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

  /** @param key The key name as produced by translateKey(). */
  handleKey(key: string): boolean {
    switch (key) {
      case 'Tab':
        // Don't allow additional popups.
        return true;
      case 'Enter':
        this.delegates.oncommit(this.resp.completions[0], this.resp.pos);
        return true;
      case 'Escape':
        this.delegates.oncommit('', this.resp.pos);
        return true;
    }
    return false;
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
  for (; pos > 0; pos--) {
    if (text.charAt(pos - 1) !== ' ') break;
  }
  for (; pos > 0; pos--) {
    if (text.charAt(pos - 1) === ' ') break;
  }

  return pos;
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

  /**
   * The selection span at time of last blur.
   * This is restored on focus, to defeat the browser behavior of
   * select all on focus.
   */
  selection: [number, number] = [0, 0];

  constructor() {
    this.prompt.innerText = '> ';
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
    switch (key) {
      case 'Delete': // At least on ChromeOS, this is M-Backspace.
      case 'M-Backspace': {
        // backward-kill-word

        const pos = this.input.selectionStart || 0;
        const start = backwardWordBoundary(this.input.value, pos);
        this.input.value =
          this.input.value.substring(0, start) +
          this.input.value.substring(pos);
        break;
      }
      case 'Enter':
        this.delegates.oncommit(this.input.value);
        break;
      case 'Tab':
        const pos = this.input.selectionStart || 0;
        const req: CompleteRequest = { input: this.input.value, pos };
        const pending = (this.pendingComplete = this.delegates.oncomplete(req));
        pending.then((resp) => {
          if (pending !== this.pendingComplete) return;
          this.pendingComplete = undefined;
          if (resp.completions.length === 0) return;
          const len = longestSharedPrefixLength(resp.completions);
          if (len > 0) {
            this.applyCompletion(
              resp.completions[0].substring(0, len),
              resp.pos
            );
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
        break;
      case 'C-a':
        this.input.selectionStart = this.input.selectionEnd = 0;
        break;
      case 'C-b':
        this.input.selectionStart = this.input.selectionEnd =
          this.input.selectionStart! - 1;
        break;
      case 'C-e':
        const len = this.input.value.length;
        this.input.selectionStart = this.input.selectionEnd = len;
        break;
      case 'C-f':
        this.input.selectionStart = this.input.selectionEnd =
          this.input.selectionStart! + 1;
        break;
      case 'C-k':
        this.input.value = this.input.value.substr(
          0,
          this.input.selectionStart!
        );
        break;
      case 'C-n':
      case 'C-p':
        // TODO: implement history.  Swallow for now.
        break;
      case 'C-u':
        this.input.value = this.input.value.substr(this.input.selectionStart!);
        break;
      case 'C-x': // browser: cut
      case 'C-c': // browser: copy
      case 'C-v': // browser: paste
      case 'C-J': // browser: inspector
      case 'C-l': // browser: location
      case 'C-R': // browser: reload
        // Allow default handling.
        return false;
      default:
        return false;
    }
    return true;
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
    this.input.value =
      this.input.value.substring(0, pos) +
      text +
      this.input.value.substring(pos + overlap);
  }
}
