import { parseIntent } from './intentParser';

// ---------------------------------------------------------------------------
// Exact phrase matches — one representative phrase per intent
// ---------------------------------------------------------------------------
describe('parseIntent – pass intent', () => {
  it('returns pass for "pass"', () => {
    expect(parseIntent('pass')).toBe('pass');
  });
  it('returns pass for "passed"', () => {
    expect(parseIntent('passed')).toBe('pass');
  });
  it('returns pass for "looks good"', () => {
    expect(parseIntent('looks good')).toBe('pass');
  });
  it('returns pass for "good"', () => {
    expect(parseIntent('good')).toBe('pass');
  });
  it('returns pass for "ok"', () => {
    expect(parseIntent('ok')).toBe('pass');
  });
  it('returns pass for "okay"', () => {
    expect(parseIntent('okay')).toBe('pass');
  });
  it('returns pass for "approve"', () => {
    expect(parseIntent('approve')).toBe('pass');
  });
  it('returns pass for "approved"', () => {
    expect(parseIntent('approved')).toBe('pass');
  });
});

describe('parseIntent – fail intent', () => {
  it('returns fail for "fail"', () => {
    expect(parseIntent('fail')).toBe('fail');
  });
  it('returns fail for "failed"', () => {
    expect(parseIntent('failed')).toBe('fail');
  });
  it('returns fail for "no good"', () => {
    expect(parseIntent('no good')).toBe('fail');
  });
  it('returns fail for "bad"', () => {
    expect(parseIntent('bad')).toBe('fail');
  });
  it('returns fail for "reject"', () => {
    expect(parseIntent('reject')).toBe('fail');
  });
  it('returns fail for "rejected"', () => {
    expect(parseIntent('rejected')).toBe('fail');
  });
  it('returns fail for "defect"', () => {
    expect(parseIntent('defect')).toBe('fail');
  });
});

describe('parseIntent – skip intent', () => {
  it('returns skip for "skip"', () => {
    expect(parseIntent('skip')).toBe('skip');
  });
  it('returns skip for "skipped"', () => {
    expect(parseIntent('skipped')).toBe('skip');
  });
  it('returns skip for "ignore"', () => {
    expect(parseIntent('ignore')).toBe('skip');
  });
  it('returns skip for "next item"', () => {
    expect(parseIntent('next item')).toBe('skip');
  });
});

describe('parseIntent – next intent', () => {
  it('returns next for "next"', () => {
    expect(parseIntent('next')).toBe('next');
  });
  it('returns next for "advance"', () => {
    expect(parseIntent('advance')).toBe('next');
  });
  it('returns next for "continue"', () => {
    expect(parseIntent('continue')).toBe('next');
  });
  it('returns next for "move on"', () => {
    expect(parseIntent('move on')).toBe('next');
  });
});

describe('parseIntent – photo intent', () => {
  it('returns photo for "photo"', () => {
    expect(parseIntent('photo')).toBe('photo');
  });
  it('returns photo for "picture"', () => {
    expect(parseIntent('picture')).toBe('photo');
  });
  it('returns photo for "image"', () => {
    expect(parseIntent('image')).toBe('photo');
  });
  it('returns photo for "camera"', () => {
    expect(parseIntent('camera')).toBe('photo');
  });
  it('returns photo for "take photo"', () => {
    expect(parseIntent('take photo')).toBe('photo');
  });
  it('returns photo for "take a photo"', () => {
    expect(parseIntent('take a photo')).toBe('photo');
  });
  it('returns photo for "take picture"', () => {
    expect(parseIntent('take picture')).toBe('photo');
  });
});

describe('parseIntent – stop intent', () => {
  it('returns stop for "stop"', () => {
    expect(parseIntent('stop')).toBe('stop');
  });
  it('returns stop for "cancel"', () => {
    expect(parseIntent('cancel')).toBe('stop');
  });
  it('returns stop for "quit"', () => {
    expect(parseIntent('quit')).toBe('stop');
  });
  it('returns stop for "exit"', () => {
    expect(parseIntent('exit')).toBe('stop');
  });
  it('returns stop for "done"', () => {
    expect(parseIntent('done')).toBe('stop');
  });
  it('returns stop for "stop listening"', () => {
    expect(parseIntent('stop listening')).toBe('stop');
  });
});

// ---------------------------------------------------------------------------
// Unknown / unrecognised input
// ---------------------------------------------------------------------------
describe('parseIntent – unknown intent', () => {
  it('returns unknown for empty string', () => {
    expect(parseIntent('')).toBe('unknown');
  });
  it('returns unknown for unrelated word', () => {
    expect(parseIntent('hello')).toBe('unknown');
  });
  it('returns unknown for random noise', () => {
    expect(parseIntent('zzz bbb qqq')).toBe('unknown');
  });
  it('returns unknown for whitespace only', () => {
    expect(parseIntent('   ')).toBe('unknown');
  });
  it('returns unknown for a number', () => {
    expect(parseIntent('42')).toBe('unknown');
  });
});

// ---------------------------------------------------------------------------
// Case insensitivity — input is uppercased / mixed-case
// ---------------------------------------------------------------------------
describe('parseIntent – case insensitivity', () => {
  it('matches PASS in uppercase', () => {
    expect(parseIntent('PASS')).toBe('pass');
  });
  it('matches FAIL in uppercase', () => {
    expect(parseIntent('FAIL')).toBe('fail');
  });
  it('matches NEXT in uppercase', () => {
    expect(parseIntent('NEXT')).toBe('next');
  });
  it('matches Skip in mixed case', () => {
    expect(parseIntent('Skip')).toBe('skip');
  });
  it('matches CAMERA in uppercase', () => {
    expect(parseIntent('CAMERA')).toBe('photo');
  });
  it('matches STOP in uppercase', () => {
    expect(parseIntent('STOP')).toBe('stop');
  });
  it('matches Looks Good in mixed case', () => {
    expect(parseIntent('Looks Good')).toBe('pass');
  });
});

// ---------------------------------------------------------------------------
// Leading / trailing whitespace handling
// ---------------------------------------------------------------------------
describe('parseIntent – whitespace trimming', () => {
  it('matches "  pass  " with surrounding spaces', () => {
    expect(parseIntent('  pass  ')).toBe('pass');
  });
  it('matches "  fail  " with surrounding spaces', () => {
    expect(parseIntent('  fail  ')).toBe('fail');
  });
  it('matches " next " with surrounding spaces', () => {
    expect(parseIntent(' next ')).toBe('next');
  });
});

// ---------------------------------------------------------------------------
// Multi-word phrase priority (longer phrases must beat shorter overlapping tokens)
// ---------------------------------------------------------------------------
describe('parseIntent – multi-word phrase priority', () => {
  // "next item" contains "next" — should resolve to skip (checked first), not next
  it('"next item" resolves to skip, not next', () => {
    expect(parseIntent('next item')).toBe('skip');
  });

  // "stop listening" contains "stop" — should resolve to stop (stop listening checked first in stop group)
  it('"stop listening" resolves to stop', () => {
    expect(parseIntent('stop listening')).toBe('stop');
  });

  // "take photo" contains "photo" — should still resolve to photo
  it('"take photo" resolves to photo', () => {
    expect(parseIntent('take photo')).toBe('photo');
  });

  // "no good" contains "good" (which maps to pass) — should resolve to fail because fail's "no good" is checked before pass's "good"
  it('"no good" resolves to fail, not pass', () => {
    expect(parseIntent('no good')).toBe('fail');
  });

  // "looks good" contains "good" — should resolve to pass via "looks good" phrase
  it('"looks good" resolves to pass', () => {
    expect(parseIntent('looks good')).toBe('pass');
  });
});

// ---------------------------------------------------------------------------
// Substring matching — phrase embedded inside a longer sentence
// ---------------------------------------------------------------------------
describe('parseIntent – substring matching in sentences', () => {
  it('detects pass intent in a natural sentence', () => {
    expect(parseIntent('that looks good to me')).toBe('pass');
  });
  it('detects fail intent in a natural sentence', () => {
    expect(parseIntent('this one looks bad')).toBe('fail');
  });
  it('detects next intent in a natural sentence containing "move on"', () => {
    // "move on" matches the 'next' intent; "next item" (→ skip) is not present
    expect(parseIntent('let us move on to the next one')).toBe('next');
  });
  it('detects photo intent in a natural sentence', () => {
    expect(parseIntent('I need to take a photo of this')).toBe('photo');
  });
  it('detects stop intent in a natural sentence', () => {
    expect(parseIntent('please stop the voice')).toBe('stop');
  });
});
