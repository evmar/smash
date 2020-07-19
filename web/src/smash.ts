import { ServerConnection } from './connection';
import { Shell } from './shell';
import { Tabs } from './tabs';
import { html, htext } from './html';

const tabs = new Tabs();

async function connect() {
  const conn = new ServerConnection();
  const hello = await conn.connect();

  const shell = new Shell();
  shell.aliases.replaceAll(
    new Map<string, string>(hello.alias.map(({ key, val }) => [key, val]))
  );
  shell.env = new Map(hello.env.map(({ key, val }) => [key, val]));
  shell.init();
  tabs.addCells(shell);
  tabs.focus();

  tabs.delegates = {
    send: (msg) => conn.send(msg),
  };

  return conn;
}

async function msgLoop(conn: ServerConnection) {
  for (;;) {
    const msg = await conn.read();
    if (!tabs.handleMessage(msg)) {
      throw new Error(`unexpected message: ${msg}`);
    }
  }
}

async function reconnectPrompt(message: string) {
  console.error(message);
  let dom!: HTMLElement;
  await new Promise((res) => {
    dom = html(
      'div',
      { className: 'error-popup' },
      html('div', {}, htext(message)),
      html('div', { style: { width: '1ex' } }),
      html(
        'button',
        {
          onclick: () => {
            res();
          },
        },
        htext('reconnect')
      )
    );
    document.body.appendChild(dom);
  });
  document.body.removeChild(dom);
}

async function main() {
  // Register an unused service worker so 'add to homescreen' works.
  // TODO: even when we do this, we still get a URL bar?!
  // await navigator.serviceWorker.register('worker.js');

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

  for (;;) {
    try {
      const conn = await connect();
      await msgLoop(conn);
    } catch (err) {
      await reconnectPrompt(err);
    }
  }
}

main().catch((err) => {
  console.error(err);
});
