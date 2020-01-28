import * as path from 'path';
import { AliasMap } from './alias';

export function parseCmd(cmd: string): string[] {
  const parts = cmd.split(/\s+/);
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

type ExecOutput = ExecRemote | TableOutput | StringOutput;

function strOutput(msg: string): ExecOutput {
  return { kind: 'string', output: msg };
}

export class Shell {
  aliases = new AliasMap();
  env = new Map<string, string>();
  cwd = '/';

  init() {
    this.cwd = this.env.get('HOME') || '/';
  }

  cwdForPrompt() {
    let cwd = this.cwd;
    const home = this.env.get('HOME');
    if (home && cwd.startsWith(home)) {
      cwd = '~' + cwd.substring(home.length);
    }
    return cwd;
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
          rows: Array.from(this.aliases.aliases)
        };
      case 'cd':
        if (argv.length > 2) {
          return strOutput('usage: cd [DIR]');
        }
        let arg = argv[1];
        if (!arg) {
          return strOutput('TODO empty cd');
        }
        if (arg.startsWith('/')) {
          arg = path.normalize(arg);
        } else {
          arg = path.join(this.cwd, arg);
        }
        return {
          kind: 'remote',
          cwd: this.cwd,
          cmd: ['cd', arg],
          onComplete: (exitCode: number) => {
            if (exitCode === 0) {
              this.cwd = arg;
            }
          }
        };
      case 'env':
        if (argv.length > 1) return;
        return {
          kind: 'table',
          headers: ['var', 'value'],
          rows: Array.from(this.env)
        };
    }
  }

  exec(cmd: string): ExecOutput {
    let argv = parseCmd(cmd);
    if (argv.length === 0) return strOutput('');
    argv = this.aliases.expand(argv);
    const out = this.handleBuiltin(argv);
    if (out) return out;
    return { kind: 'remote', cwd: this.cwd, cmd: argv };
  }
}
