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
  dom = html('div', { className: 'popup', style: { display: 'none' } });
  oncommit: (text: string, pos: number) => void = () => {};

  constructor(readonly req: CompleteRequest, readonly resp: CompleteResponse) {}

  show(parent: HTMLElement) {
    console.log(this.req, this.resp);
    const measureText = this.req.input.substring(0, this.resp.pos) + '\u200b';
    const measure = html(
      'div',
      {
        style: { position: 'absolute', visibility: 'hidden', whiteSpace: 'pre' }
      },
      htext(measureText)
    );
    parent.appendChild(measure);
    let { width, height } = getComputedStyle(measure);
    parent.removeChild(measure);

    const inputPaddingLeft = 2;
    const popupPaddingLeft = 4;
    this.dom.style.top = parseFloat(height) + 4 + 'px';
    this.dom.style.left =
      parseFloat(width) + inputPaddingLeft - popupPaddingLeft + 'px';
    this.dom.innerText = this.resp.completions.join('\n');
    this.dom.style.display = 'block';

    parent.appendChild(this.dom);
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
        this.oncommit(this.resp.completions[0], this.resp.pos);
        return true;
      case 'Escape':
        this.oncommit('', this.resp.pos);
        return true;
    }
    return false;
  }
}

export class ReadLine {
  dom = html('div', { className: 'readline' });
  prompt = html('div', { className: 'prompt' });
  inputBox = html('div', { className: 'input-box' });
  input = html('input', {
    spellcheck: false
  }) as HTMLInputElement;
  oncommit = (_: string) => {};

  oncomplete: (
    req: CompleteRequest
  ) => Promise<CompleteResponse> = async () => {
    throw 'notimpl';
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

    this.input.onkeydown = ev => {
      const key = translateKey(ev);
      if (!key) return;
      if (this.handleKey(key)) ev.preventDefault();
    };
    this.input.onkeypress = ev => {
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
      case 'Enter':
        this.oncommit(this.input.value);
        break;
      case 'Tab':
        const pos = this.input.selectionStart || 0;
        const req: CompleteRequest = { input: this.input.value, pos };
        const pending = (this.pendingComplete = this.oncomplete(req));
        pending.then(resp => {
          if (pending !== this.pendingComplete) return;
          this.popup = new CompletePopup(req, resp);
          this.popup.show(this.inputBox);
          this.popup.oncommit = (text: string, pos: number) => {
            // The completion for a partial input may include
            // some of that partial input.  Elide any text from
            // the completion that already exists in the input
            // at that same position.
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
            this.hidePopup();
          };
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
      case 'C-r': // browser: reload
      case 'C-R': // browser: reload
        // Allow default handling.
        return false;
      default:
        return false;
    }
    return true;
  }
}
