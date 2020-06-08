import { CellStack } from './cells';
import { ServerConnection } from './connection';
import * as proto from './proto';
import { Tabs } from './tabs';
import { Shell } from './shell';

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

  const conn = new ServerConnection();
  const tabs = new Tabs();

  tabs.delegates = {
    send: (msg) => conn.send(msg),
  };

  conn.delegates = {
    connect: (hello) => {
      const shell = new Shell();
      shell.aliases.setAliases(
        new Map<string, string>(hello.alias.map(({ key, val }) => [key, val]))
      );
      shell.env = new Map(hello.env.map(({ key, val }) => [key, val]));
      shell.init();
      tabs.addCells(shell);
    },

    message: (msg) => {
      if (tabs.handleMessage(msg)) return;
      console.error('unexpected message', msg);
    },
  };

  await conn.connect();

  document.body.appendChild(tabs.dom);

  // Clicking on the page, if it tries to focus the document body,
  // should redirect focus to the relevant place in the cell stack.
  // This approach feels hacky but I experimented with focus events
  // and couldn't get the desired behavior.
  document.addEventListener('click', () => {
    if (document.activeElement === document.body) {
      tabs.focus();
    }
  });
}

main().catch((err) => {
  console.error(err);
});
