/**
 * WebSocket connection setup and UI for reconnecting.
 */

import * as pb from './smash_pb';
import * as shell from './shell';
import { html } from './html';

/** Parses a WebSocket MessageEvent as a server-sent message. */
function parseMessage(event: MessageEvent): pb.ServerMsg {
  return pb.ServerMsg.deserializeBinary(new Uint8Array(event.data));
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
async function handshake(ws: WebSocket): Promise<pb.Hello> {
  const msg = await new Promise<pb.ServerMsg>((res, rej) => {
    ws.onmessage = (event) => {
      res(parseMessage(event));
    };
    ws.onclose = (event) => {
      let msg = 'connection closed';
      if (event.reason) msg += `: ${event.reason}`;
      rej(msg);
    };
  });

  const hello = msg.getHello();
  if (!hello) throw new Error(`expected hello message, got ${msg.toObject()}`);
  return hello;
}

async function connectAndHandshake(): Promise<{ws: WebSocket, hello: pb.Hello}> {
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
    connect: (msg: pb.Hello) => {},
    message: (msg: pb.ServerMsg) => {},
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

  send(msg: pb.ClientMessage): boolean {
    if (!this.ws) return false;
    this.ws.send(msg.serializeBinary());
    return true;
  }

  spawn(id: number, cmd: shell.ExecRemote): boolean {
    const run = new pb.RunRequest();
    run.setCell(id);
    run.setCwd(cmd.cwd);
    run.setArgvList(cmd.cmd);
    const msg = new pb.ClientMessage();
    msg.setRun(run);
    return this.send(msg);
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