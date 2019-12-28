/**
 * Entry point for JS-only widget interaction demo.
 */

import { html } from './html';
import { ReadLine } from './readline';

function main() {
  const readline = new ReadLine();
  document.body.appendChild(readline.dom);
}

main();
