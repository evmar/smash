import { parseCmd } from './shell';

export class AliasMap {
  aliases = new Map<string, string>();

  setAliases(aliases: Map<string, string>) {
    this.aliases = aliases;
  }

  expand(cmd: string): string {
    const first = cmd.split(' ')[0];
    let exp = this.aliases.get(first);
    if (!exp) return cmd;
    return exp + cmd.substring(first.length);
  }

  dump(): string {
    return Array.from(this.aliases.entries())
      .map(([k, v]) => `alias ${k}='${v}'\n`)
      .join('');
  }
}
