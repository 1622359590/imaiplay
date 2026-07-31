import assert from 'node:assert/strict';
import test from 'node:test';
import {
  clearAuthSession,
  isLearnerSessionToken,
  SESSION_EXPIRED_EVENT,
  TOKEN_KEY,
} from '../src/api/authSession.ts';

function token(payload: Record<string, unknown>): string {
  return `header.${Buffer.from(JSON.stringify(payload)).toString('base64url')}.signature`;
}

test('accepts only an unexpired learner token', () => {
  const nowSeconds = 2_000_000_000;

  assert.equal(
    isLearnerSessionToken(
      token({ role: 'learner', exp: nowSeconds + 60 }),
      nowSeconds * 1000,
    ),
    true,
  );
  assert.equal(
    isLearnerSessionToken(
      token({ role: 'tenant_admin', exp: nowSeconds + 60 }),
      nowSeconds * 1000,
    ),
    false,
  );
  assert.equal(
    isLearnerSessionToken(
      token({ role: 'learner', exp: nowSeconds }),
      nowSeconds * 1000,
    ),
    false,
  );
  assert.equal(isLearnerSessionToken('not-a-jwt', nowSeconds * 1000), false);
});

test('clears the token and notifies the app when a session expires', () => {
  const removed: string[] = [];
  const events: string[] = [];

  clearAuthSession(
    { removeItem: (key) => removed.push(key) },
    { dispatchEvent: (event) => events.push(event.type) },
  );

  assert.deepEqual(removed, [TOKEN_KEY]);
  assert.deepEqual(events, [SESSION_EXPIRED_EVENT]);
});
