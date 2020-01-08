import * as path from 'path';
import { AliasMap } from './alias';

function parseCmd(cmd: string): string[] {
  return cmd.split(/\s+/);
}

export interface ExecRemote {
  cwd: string;
  cmd: string[];
  onComplete?: (exitCode: number) => void;
}

export interface ExecLocal {
  output: string;
}

export function isLocal(cmd: ExecLocal | ExecRemote): cmd is ExecLocal {
  return 'output' in cmd;
}

export class Shell {
  aliases = new AliasMap();
  cwd = '/';

  exec(cmd: string): ExecLocal | ExecRemote {
    let argv = parseCmd(cmd);
    if (argv.length === 0) return { output: '' };
    argv = this.aliases.expand(argv);
    switch (argv[0]) {
      case 'cd':
        if (argv.length > 2) {
          return { output: 'usage: cd [DIR]' };
        }
        let arg = argv[1];
        if (!arg) {
          return { output: 'TODO empty cd' };
        }
        if (arg.startsWith('/')) {
          arg = path.normalize(arg);
        } else {
          arg = path.join(this.cwd, arg);
        }
        return {
          cwd: this.cwd,
          cmd: ['cd', arg],
          onComplete: (exitCode: number) => {
            if (exitCode === 0) {
              this.cwd = arg;
            }
          }
        };
      default:
        return { cwd: this.cwd, cmd: argv };
    }
  }
}
