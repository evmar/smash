/**
 * WebSocket connection setup.
 */

import * as proto from './proto';

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

async function read(ws: WebSocket): Promise<proto.ServerMsg> {
  return new Promise((res, rej) => {
    ws.onmessage = (event) => {
      res(parseMessage(event));
    };
    ws.onclose = (event) => {
      let msg = 'connection closed';
      if (event.reason) msg += `: ${event.reason}`;
      rej(msg);
    };
    ws.onerror = () => {
      // Note: it's intentional for WebSocket security reasons that you cannot
      // get much information out of a connection failure, so ignore any data
      // passe in to onerror().
      rej(`websocket connection failed`);
    };
  });
}

export class ServerConnection {
  ws!: WebSocket;

  /**
   * Opens the connection to the server.
   */
  async connect(): Promise<proto.Hello> {
    const url = new URL('/ws', window.location.href);
    url.protocol = url.protocol.replace('http', 'ws');
    const ws = new WebSocket(url.href);
    ws.binaryType = 'arraybuffer';
    await connect(ws);
    ws.onopen = (event) => {
      console.error(`unexpected ws open:`, event);
    };
    this.ws = ws;

    const msg = await read(ws);
    if (!(msg.alt instanceof proto.Hello)) {
      throw new Error(`expected hello message, got ${msg}`);
    }
    return msg.alt;
  }

  read(): Promise<proto.ServerMsg> {
    const msg = read(this.ws);
    if (TRACE_MESSAGES) console.log('<', msg);
    return msg;
  }

  send(msg: proto.ClientMessage) {
    if (TRACE_MESSAGES) console.log('>', msg);

    // Write once with an empty buffer to measure, then a second time after
    // creating the buffer.
    const writer = new proto.Writer();
    writer.writeClientMessage(msg);
    writer.buf = new Uint8Array(writer.ofs);
    writer.ofs = 0;
    writer.writeClientMessage(msg);
    this.ws.send(writer.buf);
  }
}
