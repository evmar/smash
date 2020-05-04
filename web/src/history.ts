/**
 * Manages the history of previously executed commands.
 */
export class History {
  private entries: string[] = [];

  add(cmd: string) {
    cmd = cmd.trim();
    // Avoid empty entries.
    if (cmd === '') return;
    // Avoid duplicate entries.
    if (this.entries.length > 0 && this.get(1) === cmd) return;
    this.entries.push(cmd);
  }

  get(ofs: number): string | undefined {
    if (ofs > this.entries.length) return;
    return this.entries[this.entries.length - ofs];
  }
}
