import { flatbuffers } from 'flatbuffers';
import { proto } from './smash_generated';

let ws: WebSocket;
const out = document.createElement('pre');

function spawn(cmd: string) {
  const builder = new flatbuffers.Builder();
  const r = proto.ReqRun.create(builder, builder.createString(cmd));
  builder.finish(r);
  ws.send(builder.asUint8Array());
}

function handleMessage(ev: MessageEvent) {
  const buf = new flatbuffers.ByteBuffer(new Uint8Array(ev.data));
  const msg = proto.RespOutput.getRoot(buf);
  out.innerText += msg.text();
}

function connect() {
  const url = new URL('/ws', window.location.href);
  url.protocol = url.protocol.replace('http', 'ws');
  ws = new WebSocket(url.href);
  ws.binaryType = 'arraybuffer';
  ws.onopen = event => {
    spawn('ls');
  };
  ws.onmessage = handleMessage;
}

function main() {
  document.body.appendChild(out);
  connect();
}

main();
