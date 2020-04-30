import * as readline from './readline';
import { expect } from 'chai';

describe.only('word boundaries', () => {
  it('backward', () => {
    expect(readline.backwardWordBoundary('', 0)).equal(0);
    expect(readline.backwardWordBoundary('ab', 1)).equal(0);

    expect(readline.backwardWordBoundary('ab cd', 5)).equal(3);
  });
});
