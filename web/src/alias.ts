import { parseCmd } from './shell';

export class AliasMap {
  aliases = new Map<string, string>();

  setAliases(aliases: Map<string, string>) {
    this.aliases = aliases;
  }

  expand(cmd: string[]): string[] {
    const exp = this.aliases.get(cmd[0]);
    if (!exp) return cmd;
    return parseCmd(exp).concat(cmd.slice(1));
  }

  dump(): string {
    return Array.from(this.aliases.entries())
      .map(([k, v]) => `alias ${k}='${v}'\n`)
      .join('');
  }
}
