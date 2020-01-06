export class AliasMap {
  aliases = new Map<string, string[]>();

  constructor() {
    this.aliases.set('ls', ['ls', '--color']);
  }

  expand(cmd: string[]): string[] {
    const exp = this.aliases.get(cmd[0]);
    if (!exp) return cmd;
    return exp.concat(cmd.slice(1));
  }
}
