/**
 * Entry point for JS-only widget interaction demo.
 */

import { ReadLine } from './readline';
import { History } from './history';

function main() {
  const readline = new ReadLine(new History());
  document.body.appendChild(readline.dom);
}

main();
