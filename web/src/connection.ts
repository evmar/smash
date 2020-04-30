/**
 * WebSocket connection setup and UI for reconnecting.
 */

import * as proto from './proto';
import { html } from './html';

const TRACE_MESSAGES = false;

/** Parses a WebSocket MessageEvent as a server-sent message. */
function parseMessage(event: MessageEvent): proto.ServerMsg {
  const msg = new proto.Reader(new DataView(event.data)).readServerMsg();
  if (TRACE_MESSAGES) console.log('<', msg);
  return msg;
}

/** Promisifies WebSocket connection. */
function connect(ws: WebSocket): Promise<void> {
  return new Promise((res, rej) => {
    ws.onopen = () => {
      res();
    };
    ws.onerror = () => {
      // Note: it's intentional for WebSocket security reasons that you cannot
      // get much information out of a connection failure, so ignore any data
      // passe in to onerror().
      rej(`websocket connection failed`);
    };
  });
}

/** Waits for the server hello message. */
async function handshake(ws: WebSocket): Promise<proto.Hello> {
  const msg = await new Promise<proto.ServerMsg>((res, rej) => {
    ws.onmessage = (event) => {
      res(parseMessage(event));
    };
    ws.onclose = (event) => {
      let msg = 'connection closed';
      if (event.reason) msg += `: ${event.reason}`;
      rej(msg);
    };
  });

  if (!(msg.alt instanceof proto.Hello))
    throw new Error(`expected hello message, got ${msg}`);
  return msg.alt;
}

async function connectAndHandshake(): Promise<{
  ws: WebSocket;
  hello: proto.Hello;
}> {
  const url = new URL('/ws', window.location.href);
  url.protocol = url.protocol.replace('http', 'ws');
  const ws = new WebSocket(url.href);
  ws.binaryType = 'arraybuffer';
  await connect(ws);
  ws.onopen = (event) => {
    console.error(`unexpected ws open:`, event);
  };
  const hello = await handshake(ws);
  return { ws, hello };
}

/** Maintains a WebSocket connection to the server, showing reconnect UI if necessary. */
export class ServerConnection {
  delegates = {
    connect: (msg: proto.Hello) => {},
    message: (msg: proto.ServerMsg) => {},
  };

  private ws: WebSocket | null = null;
  private dom: HTMLElement | null = null;

  /**
   * Opens the connection, and resolves after the first connect.
   * Attempts to keep it open by showing reconnect UI.
   */
  async connect(): Promise<void> {
    this.removeDOM();
    this.ws = null;

    try {
      const { ws, hello } = await connectAndHandshake();
      ws.onmessage = (event) => {
        this.delegates.message(parseMessage(event));
      };
      ws.onclose = (event) => {
        let msg = 'connection closed';
        if (event.reason) msg += `: ${event.reason}`;
        this.showError(msg);
      };
      this.ws = ws;
      this.delegates.connect(hello);
    } catch (err) {
      this.showError(err);
    }
  }

  send(msg: proto.ClientMessage): boolean {
    if (TRACE_MESSAGES) console.log('>', msg);
    if (!this.ws) return false;
    // Write once with an empty buffer to measure, then a second time after
    // creating the buffer.
    const writer = new proto.Writer();
    writer.writeClientMessage(msg);
    writer.buf = new Uint8Array(writer.ofs);
    writer.ofs = 0;
    writer.writeClientMessage(msg);
    this.ws.send(writer.buf);
    return true;
  }

  private removeDOM() {
    if (!this.dom) return;
    this.dom.parentElement!.removeChild(this.dom);
    this.dom = null;
  }

  private showError(msg: string) {
    this.ws = null;
    console.error(msg);
    this.removeDOM();
    this.dom = html(
      'div',
      { className: 'error-popup' },
      html('div', {}, document.createTextNode(msg)),
      html('div', { style: { width: '1ex' } }),
      html(
        'button',
        {
          onclick: () => {
            this.connect();
          },
        },
        document.createTextNode('reconnect')
      )
    );
    document.body.appendChild(this.dom);
  }
}
