import { CellStack } from './cells';
import { ServerConnection } from './connection';
import * as proto from './proto';

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

  const conn = new ServerConnection();
  const cellStack = new CellStack();

  cellStack.delegates = {
    send: (msg) => conn.send(msg),
  };

  conn.delegates = {
    connect: (hello) => {
      const shell = cellStack.shell;
      shell.aliases.setAliases(
        new Map<string, string>(hello.alias.map(({ key, val }) => [key, val]))
      );
      shell.env = new Map(hello.env.map(({ key, val }) => [key, val]));
      shell.init();
    },

    message: (msg) => {
      if (msg.alt instanceof proto.CompleteResponse) {
        cellStack.getLastCell().onCompleteResponse(msg.alt);
      } else if (msg.alt instanceof proto.CellOutput) {
        cellStack.onOutput(msg.alt);
      } else {
        console.error('unexpected message', msg);
      }
    },
  };

  await conn.connect();

  cellStack.addNew();

  // Clicking on the page, if it tries to focus the document body,
  // should redirect focus to the relevant place in the cell stack.
  // This approach feels hacky but I experimented with focus events
  // and couldn't get the desired behavior.
  document.addEventListener('click', () => {
    if (document.activeElement === document.body) {
      cellStack.focus();
    }
  });
}

main().catch((err) => {
  console.error(err);
});
