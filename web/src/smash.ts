import * as pb from './smash_pb';

let ws: WebSocket | null = null;
const out = document.createElement('pre');

function div(className: string) {
  const div = document.createElement('div');
  div.className = className;
  return div;
}

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
  if (name.length === 0) return '';
  name += ev.key;
  return name;
}

class ReadLine {
  dom = div('readline');
  prompt = div('prompt');
  input = document.createElement('input');
  oncommit = (_: string) => {};

  constructor() {
    this.prompt.innerText = '> ';
    this.dom.appendChild(this.prompt);

    this.input.className = 'input';
    this.input.spellcheck = false;
    this.dom.appendChild(this.input);

    this.input.onkeydown = ev => {
      this.keydown(ev);
    };
    this.input.onkeypress = ev => {
      this.keypress(ev);
    };
  }

  keydown(ev: KeyboardEvent) {
    const key = translateKey(ev);
    if (!key) return;
    switch (key) {
      case 'C-a':
        this.input.selectionStart = this.input.selectionEnd = 0;
        break;
      case 'C-e':
        const len = this.input.value.length;
        this.input.selectionStart = this.input.selectionEnd = len;
        break;
      case 'C-k':
        this.input.value = this.input.value.substr(
          0,
          this.input.selectionStart!
        );
        break;
      case 'C-J': // browser: inspector
      case 'C-l': // browser: location
      case 'C-r': // browser: reload
        // Allow default handling.
        return;
      default:
        console.log('TODO:', key, ev);
    }
    ev.preventDefault();
  }

  keypress(ev: KeyboardEvent) {
    switch (ev.key) {
      case 'Enter':
        this.oncommit(this.input.value);
        break;
      default:
        return;
    }
    ev.preventDefault();
  }
}

function spawn(cmd: string) {
  if (!ws) return;
  const msg = new pb.RunRequest();
  msg.setCommand(cmd);
  ws.send(msg.serializeBinary());
}

function handleMessage(ev: MessageEvent) {
  const msg = pb.OutputResponse.deserializeBinary(new Uint8Array(ev.data));
  out.innerText += msg.getText();
}

async function connect(): Promise<WebSocket> {
  const url = new URL('/ws', window.location.href);
  url.protocol = url.protocol.replace('http', 'ws');
  const ws = new WebSocket(url.href);
  ws.binaryType = 'arraybuffer';
  ws.onmessage = handleMessage;
  return new Promise((res, rej) => {
    ws.onopen = () => {
      res(ws);
    };
    ws.onerror = err => {
      rej(err);
    };
  });
}

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

  const rl = new ReadLine();
  document.body.appendChild(rl.dom);
  rl.input.focus();
  document.body.appendChild(out);
  ws = await connect();
  ws.onclose = ev => {
    out.innerText += `\nconnection closed: ${ev.code} (${ev.reason})\n`;
    ws = null;
  };
  ws.onerror = err => {
    out.innerText += `\nconnection failed: ${err}\n`;
    ws = null;
  };

  rl.oncommit = cmd => {
    rl.input.blur();
    spawn(cmd);
  };
}

main().catch(err => {
  console.error(err);
});
