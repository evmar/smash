/**
 * Replacements for node's "path" module.
 */

export function join(a: string, b: string): string {
  if (b.startsWith('/')) return b;
  if (a.endsWith('/')) return a;
  return normalize(`${a}/${b}`);
}

export function normalize(path: string): string {
  const partsIn = path.split('/');
  const partsOut: string[] = [];
  for (const part of partsIn) {
    switch (part) {
      case '':
        if (partsOut.length > 0) continue;
        break;
      case '.':
        if (partsOut.length > 0) continue;
        break;
      case '..':
        if (partsOut.length > 0) {
          partsOut.pop();
          continue;
        }
        break;
    }
    partsOut.push(part);
  }
  return partsOut.join('/');
}
