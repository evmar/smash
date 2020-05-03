/**
 * Manages the history of previously executed commands.
 */
export class History {
  private entries: string[] = [];

  add(cmd: string) {
    if (cmd.trim() === '') return;
    this.entries.push(cmd);
  }

  get(ofs: number): string | undefined {
    if (ofs > this.entries.length) return;
    return this.entries[this.entries.length - ofs];
  }
}
