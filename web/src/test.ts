import { ReadLine } from './readline';
import { expect } from 'chai';
import * as http from 'http';
import * as puppeteer from 'puppeteer';

import { runServer, port } from './server';

let server: http.Server;
let browser: puppeteer.Browser;

before(async () => {
  server = await runServer();
  browser = await puppeteer.launch({
    // headless: false,
    // slowMo: 500,
  });
});

declare const smash: typeof import('./widgets').exported;

after(async () => {
  await browser.close();
  server.close();
});

describe('readline', async function () {
  let page: puppeteer.Page;
  let readline: puppeteer.JSHandle;

  beforeEach(async () => {
    page = await browser.newPage();
    await page.goto(`http://localhost:${port}/test.html`);
    readline = await page.evaluateHandle(() => {
      const readline = new smash.ReadLine();
      document.body.appendChild(readline.dom);
      readline.input.focus();
      return readline;
    });
  });

  function getCursorPos() {
    return page.evaluate(
      (readline: ReadLine) => readline.input.selectionStart,
      readline
    );
  }

  describe('emacs', () => {
    async function typeEmacs(key: string) {
      let control = false;
      if (key.startsWith('C-')) {
        control = true;
        key = key.substr(2);
      }

      if (control) await page.keyboard.down('Control');
      await page.keyboard.type(key);
      if (control) await page.keyboard.up('Control');
    }

    it('C-a', async () => {
      await page.keyboard.type('demo');
      expect(await getCursorPos()).equal('demo'.length);
      await typeEmacs('C-a');
      expect(await getCursorPos()).equal(0);
    });
  });
});
