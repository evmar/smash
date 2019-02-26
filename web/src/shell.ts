import * as path from 'path';

function parseCmd(cmd: string): string[] {
  return cmd.split(/\s+/);
}

export interface ExecRemote {
  cwd: string;
  cmd: string[];
}

export interface ExecLocal {
  output: string;
}

export function isLocal(cmd: ExecLocal | ExecRemote): cmd is ExecLocal {
  return 'output' in cmd;
}

export class Shell {
  cwd = '/';

  exec(cmd: string): ExecLocal | ExecRemote {
    const argv = parseCmd(cmd);
    if (argv.length === 0) return { output: '' };
    switch (argv[0]) {
      case 'cd':
        if (argv.length > 2) {
          return { output: 'usage: cd [DIR]' };
        } else if (argv.length === 1) {
          return { output: '' };
        } else {
          const arg = argv[1];
          if (arg.startsWith('/')) {
            this.cwd = path.normalize(arg);
          } else {
            this.cwd = path.join(this.cwd, arg);
          }
          return { output: '' };
        }
      default:
        return { cwd: this.cwd, cmd: argv };
    }
  }
}
