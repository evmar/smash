import * as path from './path';
import { expect } from 'chai';

describe('path', () => {
  describe('join', () => {
    it('joins relatively', () => {
      expect(path.join('a', 'b')).equal('a/b');
    });

    it('obeys parents', () => {
      expect(path.join('a/b', '../c')).equal('a/c');
    });
  });

  describe('normalize', () => {
    it('leaves ok paths alone', () => {
      expect(path.normalize('/foo/bar')).equal('/foo/bar');
      expect(path.normalize('foo/bar')).equal('foo/bar');
    });

    it('normalizes double slash', () => {
      expect(path.normalize('a//b')).equal('a/b');
    });

    it('normalizes dot', () => {
      expect(path.normalize('./bar')).equal('./bar');
      expect(path.normalize('foo/./bar')).equal('foo/bar');
    });
  });
});
