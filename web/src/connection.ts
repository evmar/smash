/**
 * WebSocket connection setup.
 */

import * as proto from './proto';

const TRACE_MESSAGES = true;

/** Prints a proto-encoded message. */
function printMessage(prefix: string, msg: any) {
  if ('alt' in msg) {
    const alt = msg.alt;
    printMessage(`${prefix}${msg.constructor.name}:`, alt);
    return;
  }
  console.groupCollapsed(`${prefix}${msg.constructor.name}`);
  for (const field in msg) {
    const val = msg[field];
    if (typeof val === 'object' && !Array.isArray(val)) {
      printMessage(`${field}: `, val);
    } else {
      console.info(`${field}:`, val);
    }
  }
  console.groupEnd();
}

/** Parses a WebSocket MessageEvent as a server-sent message. */
function parseMessage(event: MessageEvent): proto.ServerMsg {
  return new proto.Reader(new DataView(event.data)).readServerMsg();
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
    if (msg.tag !== 'Hello') {
      throw new Error(`expected hello message, got ${msg}`);
    }
    return msg.val;
  }

  async read(): Promise<proto.ServerMsg> {
    const msg = await read(this.ws);
    if (TRACE_MESSAGES) printMessage('<', msg);
    return msg;
  }

  send(msg: proto.ClientMessage) {
    if (TRACE_MESSAGES) printMessage('>', msg);

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
