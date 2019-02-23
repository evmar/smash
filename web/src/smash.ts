import * as pb from './smash_pb';

let ws: WebSocket;
const out = document.createElement('pre');

function spawn(cmd: string) {
  const msg = new pb.RunRequest();
  msg.setCommand(cmd);
  ws.send(msg.serializeBinary());
}

function handleMessage(ev: MessageEvent) {
  const msg = pb.OutputResponse.deserializeBinary(new Uint8Array(ev.data));
  out.innerText += msg.getText();
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
