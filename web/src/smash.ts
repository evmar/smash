console.log('hello, world');

const url = new URL('/ws', window.location.href);
url.protocol = url.protocol.replace('http', 'ws');
const ws = new WebSocket(url.href);
ws.onopen = event => {
  ws.send('test message');
};
