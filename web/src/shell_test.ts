import { Shell, ExecOutput } from './shell';
import { expect } from 'chai';

async function fakeExec(out: ExecOutput): Promise<void> {
  if (out.kind !== 'remote') return;
  return new Promise((res, rej) => {
    out.onComplete?.(0);
    res();
  });
}

describe('shell', async function() {
  const env = new Map<string, string>([['HOME', '/home/evmar']]);

  it('elides homedir', function() {
    const sh = new Shell(env);
    sh.cwd = '/home/evmar';
    expect(sh.cwdForPrompt()).equal('~');
    sh.cwd = '/home/evmar/test';
    expect(sh.cwdForPrompt()).equal('~/test');
  });

  describe('cd', function() {
    it('goes home', async function() {
      const sh = new Shell(env);
      await fakeExec(sh.builtinCd([]));
      expect(sh.cwd).equal('/home/evmar');
    });

    it('normalizes paths', async function() {
      const sh = new Shell(env);
      await fakeExec(sh.builtinCd([]));
      await fakeExec(sh.builtinCd(['foo//bar/']));
      expect(sh.cwd).equal('/home/evmar/foo/bar');
    });
  });
});
