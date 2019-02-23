import { flatbuffers } from 'flatbuffers';
import { proto } from './smash_generated';

const url = new URL('/ws', window.location.href);
url.protocol = url.protocol.replace('http', 'ws');
const ws = new WebSocket(url.href);
ws.onopen = event => {
  const b = new flatbuffers.Builder();
  const r = proto.ReqRun.create(b, b.createString('test message'));
  b.finish(r);
  ws.send(b.asUint8Array());
};
