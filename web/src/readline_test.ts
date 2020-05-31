import * as readline from './readline';
import { expect } from 'chai';
import { cursorTo } from 'readline';

function cursor(text: string): [string, number] {
  const pos = text.indexOf('|');
  if (pos === -1) return [text, 0];
  return [text.substring(0, pos) + text.substring(pos + 1), pos];
}

class Fake implements readline.InputHandler {
  text = '';
  pos = 0;
  history = 0;

  onEnter() {}
  tabComplete(state: {}): void {}
  setText(text: string): void {
    this.text = text;
  }
  setPos(pos: number): void {
    this.pos = pos;
  }
  showHistory(delta: -1 | 0 | 1): void {
    this.history = delta;
  }

  set(state: string) {
    [this.text, this.pos] = cursor(state);
  }


  interpret(key: string) {
    readline.interpretKey({text: this.text, start: this.pos, end: this.pos}, key, this);
  }
  expect(newState: string) {
    const [etext, epos] = cursor(newState);
    expect(this.text).equal(etext);
    expect(this.pos).equal(epos);
  }
}

describe('readline', () => {
  describe('word boundaries', () => {
    it('backward', () => {
      function expectBack(from: string, to: string) {
        const [text, pos1] = cursor(from);
        const [, pos2] = cursor(to);
        expect(readline.backwardWordBoundary(text, pos1)).equal(pos2);
      }
      expectBack('', '');
      expectBack('a|b', '|ab');
      expectBack('ab cd|', 'ab |cd');
    });
  });

  describe('interpretKey', () => {
    it('basic movement', () => {
      const fake = new Fake();
      fake.set('hello world|');
      fake.interpret('C-b');
      fake.expect('hello worl|d');
      fake.interpret('M-b');
      fake.expect('hello |world');
      fake.interpret('C-a');
      fake.expect('|hello world');
      fake.interpret('M-f');
      fake.expect('hello |world');
      fake.interpret('C-e');
      fake.expect('hello world|');
    });

    it('history', () => {
      const fake = new Fake();
      fake.interpret('C-p');
      expect(fake.history).equal(1);

      // Hitting C-c shouldn't reset history state.
      fake.interpret('C-c');
      expect(fake.history).equal(1);

      // Normal typing should reset history state.
      fake.interpret('c');
      expect(fake.history).equal(0);

      // Also arrow keys.
      fake.interpret('ArrowUp');
      expect(fake.history).equal(1);
      fake.interpret('ArrowDown');
      expect(fake.history).equal(-1);
    });
  });
});
