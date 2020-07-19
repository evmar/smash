import { AliasMap } from './alias';
import * as path from './path';

export function parseCmd(cmd: string): string[] {
  const parts = cmd.trim().split(/\s+/);
  if (parts.length === 1 && parts[0] === '') return [];
  return parts;
}

export interface ExecRemote {
  kind: 'remote';
  cwd: string;
  cmd: string[];
  onComplete?: (exitCode: number) => void;
}

export interface TableOutput {
  kind: 'table';
  headers: string[];
  rows: string[][];
}

export interface StringOutput {
  kind: 'string';
  output: string;
}

export type ExecOutput = ExecRemote | TableOutput | StringOutput;

function strOutput(msg: string): ExecOutput {
  return { kind: 'string', output: msg };
}

export class Shell {
  aliases = new AliasMap();
  cwd = '/';

  constructor(public env = new Map<string, string>()) {}

  init() {
    this.cwd = this.env.get('HOME') || '/';
    this.aliases.set('that', `${this.env.get('SMASH')} that`);
  }

  cwdForPrompt() {
    let cwd = this.cwd;
    const home = this.env.get('HOME');
    if (home && cwd.startsWith(home)) {
      cwd = '~' + cwd.substring(home.length);
    }
    return cwd;
  }

  builtinCd(argv: string[]): ExecOutput {
    if (argv.length > 1) {
      return strOutput('usage: cd [DIR]');
    }
    let arg = argv[0];
    if (!arg) {
      arg = this.env.get('HOME') || '/';
    }
    if (!arg.startsWith('/')) {
      arg = path.join(this.cwd, arg);
    }
    arg = path.normalize(arg);
    if (arg.length > 1 && arg.endsWith('/')) {
      arg = arg.substring(0, arg.length - 1);
    }
    return {
      kind: 'remote',
      cwd: this.cwd,
      cmd: ['cd', arg],
      onComplete: (exitCode: number) => {
        if (exitCode === 0) {
          this.cwd = arg;
        }
      },
    };
  }

  private handleBuiltin(argv: string[]): ExecOutput | undefined {
    switch (argv[0]) {
      case 'alias':
        if (argv.length > 2) {
          return strOutput('usage: alias [CMD]');
        }
        if (argv.length > 1) {
          return strOutput('TODO: alias CMD');
        }
        return {
          kind: 'table',
          headers: ['alias', 'expansion'],
          rows: Array.from(this.aliases.aliases),
        };
      case 'cd':
        return this.builtinCd(argv.slice(1));
      case 'env':
        if (argv.length > 1) return;
        return {
          kind: 'table',
          headers: ['var', 'value'],
          rows: Array.from(this.env),
        };
    }
  }

  exec(cmd: string): ExecOutput {
    cmd = cmd.trim();
    cmd = this.aliases.expand(cmd);
    const argv = parseCmd(cmd);
    const out = this.handleBuiltin(argv);
    if (out) return out;
    return { kind: 'remote', cwd: this.cwd, cmd: ['/bin/sh', '-c', cmd] };
  }
}
