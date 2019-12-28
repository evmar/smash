/**
 * Basic web server, used in tests and in local dev.
 */

import * as fs from 'fs';
import * as http from 'http';
import * as path from 'path';
import * as url from 'url';

export const port = 9001;

export function runServer(): Promise<http.Server> {
  const server = http.createServer((req, res) => {
    const reqUrl = url.parse(req.url || '/');
    let reqPath = path.normalize(reqUrl.path || '/');
    if (reqPath.endsWith('/')) reqPath += 'index.html';
    if (!reqPath.startsWith('/')) {
      console.error('bad request', reqPath);
      throw new Error('bad request');
    }
    reqPath = path.join('dist', reqPath);

    const file = fs.createReadStream(reqPath);
    file.pipe(res);
  });
  server.listen(port);
  return new Promise(resolve => {
    server.on('listening', () => {
      console.log(`test server listening on ${port}`);
      resolve(server);
    });
  });
}

if (require.main === module) {
  runServer();
}
