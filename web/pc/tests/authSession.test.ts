import assert from 'node:assert/strict';
import test from 'node:test';
import * as authSession from '../src/api/authSession.ts';

const sessionApi = authSession as typeof authSession & {
  PORTAL_ACCESS_TOKEN_KEY?: string;
  PORTAL_REFRESH_TOKEN_KEY?: string;
  PORTAL_TENANT_CODE_KEY?: string;
  isPortalSessionToken?: (token: string | null, tenantId?: string, now?: number) => boolean;
  migrateLegacySession?: (storage: StorageLike, now?: number) => void;
  bindPortalSessionToPortal?: (
    portal: { code: string; tenant_id: string }, storage: StorageLike, now?: number,
  ) => boolean;
  writeAdminSession?: (
    session: { token: string; refresh_token?: string }, storage: StorageLike,
  ) => void;
  validateAuthenticatedSession?: (result: unknown) => void;
  readPortalTenantCode?: (storage: Pick<StorageLike, 'getItem'>) => string | undefined;
  createPortalLogoutRequest?: (
    storage: Pick<StorageLike, 'getItem'>,
  ) => { refreshToken: string; authorization: string } | undefined;
  createPortalSessionRefresher?: (
    request: (
      refreshToken: string,
      portal: { code: string; tenant_id: string },
    ) => Promise<{ token: string; refresh_token?: string }>,
    currentPortal: () => { code: string; tenant_id: string } | undefined,
    storage: StorageLike,
  ) => () => Promise<string>;
  shouldRefreshPortalRequest?: (input: {
    status?: number;
    url?: string;
    retried?: boolean;
    hasRefreshToken: boolean;
  }) => boolean;
  readValidLegacyStaffRole?: (
    storage: Pick<StorageLike, 'getItem'>,
    now?: number,
  ) => 'instructor' | 'tenant_admin' | 'superadmin' | undefined;
};

interface StorageLike {
  getItem: (key: string) => string | null;
  setItem: (key: string, value: string) => void;
  removeItem: (key: string) => void;
}

function token(payload: Record<string, unknown>): string {
  return `header.${Buffer.from(JSON.stringify(payload)).toString('base64url')}.signature`;
}

function memoryStorage(initial: Record<string, string> = {}): StorageLike & { value: (key: string) => string | undefined } {
  const entries = new Map(Object.entries(initial));
  return {
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => entries.set(key, value),
    removeItem: (key) => entries.delete(key),
    value: (key) => entries.get(key),
  };
}

test('accepts only an unexpired learner token', () => {
  const nowSeconds = 2_000_000_000;

  assert.equal(
    authSession.isLearnerSessionToken(
      token({ role: 'learner', exp: nowSeconds + 60 }),
      nowSeconds * 1000,
    ),
    true,
  );
  assert.equal(
    authSession.isLearnerSessionToken(
      token({ role: 'tenant_admin', exp: nowSeconds + 60 }),
      nowSeconds * 1000,
    ),
    false,
  );
  assert.equal(
    authSession.isLearnerSessionToken(
      token({ role: 'learner', exp: nowSeconds }),
      nowSeconds * 1000,
    ),
    false,
  );
  assert.equal(authSession.isLearnerSessionToken('not-a-jwt', nowSeconds * 1000), false);
});

test('clears the token and notifies the app when a session expires', () => {
  const removed: string[] = [];
  const events: string[] = [];

  authSession.clearAuthSession(
    { removeItem: (key) => removed.push(key) },
    { dispatchEvent: (event) => events.push(event.type) },
  );

  assert.deepEqual(removed, [
    sessionApi.PORTAL_ACCESS_TOKEN_KEY,
    sessionApi.PORTAL_REFRESH_TOKEN_KEY,
    sessionApi.PORTAL_TENANT_CODE_KEY,
  ]);
  assert.deepEqual(events, [authSession.SESSION_EXPIRED_EVENT]);
});

test('migrates an unexpired legacy learner token to the portal key', () => {
  const entries = new Map<string, string>();
  const storage = {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => entries.set(key, value),
    removeItem: (key: string) => entries.delete(key),
  };
  const learnerToken = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  storage.setItem(authSession.TOKEN_KEY, learnerToken);

  assert.equal(typeof sessionApi.migrateLegacySession, 'function');
  if (!sessionApi.migrateLegacySession) return;
  sessionApi.migrateLegacySession(storage, 2_000_000_000 * 1000);

  assert.equal(storage.getItem(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), learnerToken);
  assert.equal(storage.getItem(authSession.TOKEN_KEY), null);
});

test('does not migrate an admin token into the portal session', () => {
  const entries = new Map<string, string>();
  const storage = {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => entries.set(key, value),
    removeItem: (key: string) => entries.delete(key),
  };
  const adminToken = token({ user_id: 'admin-1', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  storage.setItem(authSession.TOKEN_KEY, adminToken);

  assert.equal(typeof sessionApi.migrateLegacySession, 'function');
  if (!sessionApi.migrateLegacySession) return;
  sessionApi.migrateLegacySession(storage, 2_000_000_000 * 1000);

  assert.equal(storage.getItem(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), null);
  assert.equal(storage.getItem(authSession.TOKEN_KEY), adminToken);
});

test('recognizes only valid staff legacy sessions for the Admin handoff', () => {
  const now = 2_000_000_000 * 1000;
  assert.equal(typeof sessionApi.readValidLegacyStaffRole, 'function');
  if (!sessionApi.readValidLegacyStaffRole) return;

  const tenantAdmin = memoryStorage({
    [authSession.TOKEN_KEY]: token({
      user_id: 'admin-1', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 2_000_000_060,
    }),
  });
  const superadmin = memoryStorage({
    [authSession.TOKEN_KEY]: token({
      user_id: 'root-1', role: 'superadmin', tenant_id: '', exp: 2_000_000_060,
    }),
  });
  const learner = memoryStorage({
    [authSession.TOKEN_KEY]: token({
      user_id: 'learner-1', role: 'learner', tenant_id: 'tenant-acme', exp: 2_000_000_060,
    }),
  });
  const expired = memoryStorage({
    [authSession.TOKEN_KEY]: token({
      user_id: 'admin-2', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 1,
    }),
  });

  assert.equal(sessionApi.readValidLegacyStaffRole(tenantAdmin, now), 'tenant_admin');
  assert.equal(sessionApi.readValidLegacyStaffRole(superadmin, now), 'superadmin');
  assert.equal(sessionApi.readValidLegacyStaffRole(learner, now), undefined);
  assert.equal(sessionApi.readValidLegacyStaffRole(expired, now), undefined);
});

test('removes a valid learner legacy token when a scoped portal session already exists', () => {
  const current = token({ user_id: 'user-current', role: 'learner', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  const legacy = token({ user_id: 'user-legacy', role: 'learner', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  const storage = memoryStorage({
    [sessionApi.PORTAL_ACCESS_TOKEN_KEY!]: current,
    [authSession.TOKEN_KEY]: legacy,
  });

  sessionApi.migrateLegacySession?.(storage, 2_000_000_000 * 1000);

  assert.equal(storage.value(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), current);
  assert.equal(storage.value(authSession.TOKEN_KEY), undefined);
});

test('keeps a staff legacy token when a scoped portal session already exists', () => {
  const current = token({ user_id: 'user-current', role: 'learner', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  const staff = token({ user_id: 'staff-legacy', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  const storage = memoryStorage({
    [sessionApi.PORTAL_ACCESS_TOKEN_KEY!]: current,
    [authSession.TOKEN_KEY]: staff,
  });

  sessionApi.migrateLegacySession?.(storage, 2_000_000_000 * 1000);

  assert.equal(storage.value(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), current);
  assert.equal(storage.value(authSession.TOKEN_KEY), staff);
});

test('requires an unexpired learner token with a matching tenant for an explicit portal', () => {
  const nowSeconds = 2_000_000_000;
  const learnerToken = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: nowSeconds + 60 });

  assert.equal(typeof sessionApi.isPortalSessionToken, 'function');
  if (!sessionApi.isPortalSessionToken) return;
  assert.equal(sessionApi.isPortalSessionToken(learnerToken, 'tenant-acme', nowSeconds * 1000), true);
  assert.equal(sessionApi.isPortalSessionToken(learnerToken, 'tenant-other', nowSeconds * 1000), false);
  assert.equal(
    sessionApi.isPortalSessionToken(token({ role: 'learner', exp: nowSeconds + 60 }), 'tenant-acme', nowSeconds * 1000),
    false,
  );
});

test('binds a legacy learner session only after the resolved portal tenant matches', () => {
  const entries = new Map<string, string>();
  const storage = {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => entries.set(key, value),
    removeItem: (key: string) => entries.delete(key),
  };
  const learnerToken = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: 2_000_000_060 });
  storage.setItem(authSession.TOKEN_KEY, learnerToken);

  assert.equal(typeof sessionApi.bindPortalSessionToPortal, 'function');
  if (!sessionApi.bindPortalSessionToPortal) return;
  assert.equal(
    sessionApi.bindPortalSessionToPortal({ code: 'acme', tenant_id: 'tenant-acme' }, storage, 2_000_000_000 * 1000),
    true,
  );
  assert.equal(storage.getItem(sessionApi.PORTAL_TENANT_CODE_KEY!), 'acme');
  assert.equal(storage.getItem(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), learnerToken);
});

test('clears a foreign portal session instead of binding it to another portal', () => {
  const entries = new Map<string, string>();
  const storage = {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => entries.set(key, value),
    removeItem: (key: string) => entries.delete(key),
  };
  storage.setItem(sessionApi.PORTAL_ACCESS_TOKEN_KEY!, token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-bravo', exp: 2_000_000_060 }));

  assert.equal(typeof sessionApi.bindPortalSessionToPortal, 'function');
  if (!sessionApi.bindPortalSessionToPortal) return;
  assert.equal(
    sessionApi.bindPortalSessionToPortal({ code: 'acme', tenant_id: 'tenant-acme' }, storage, 2_000_000_000 * 1000),
    false,
  );
  assert.equal(storage.getItem(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), null);
  assert.equal(storage.getItem(sessionApi.PORTAL_TENANT_CODE_KEY!), null);
});

test('reads the persistent portal binding after platform routing clears transient state', () => {
  const storage = {
    getItem: (key: string) => key === sessionApi.PORTAL_TENANT_CODE_KEY ? '  Acme  ' : null,
  };
  assert.equal(typeof sessionApi.readPortalTenantCode, 'function');
  if (!sessionApi.readPortalTenantCode) return;
  assert.equal(sessionApi.readPortalTenantCode(storage), 'acme');
});

test('writes a tenantless superadmin session but rejects expired staff sessions', () => {
  const entries = new Map<string, string>();
  const storage = {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => entries.set(key, value),
    removeItem: (key: string) => entries.delete(key),
  };
  assert.equal(typeof sessionApi.writeAdminSession, 'function');
  if (!sessionApi.writeAdminSession) return;

  const superadminToken = token({ user_id: 'root-1', role: 'superadmin', tenant_id: '', exp: 2_000_000_060 });
  assert.doesNotThrow(() => sessionApi.writeAdminSession!({ token: superadminToken }, storage));
  assert.throws(() => sessionApi.writeAdminSession!({ token: token({ user_id: 'staff-1', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 1 }) }, storage));
});

test('accepts a tenantless superadmin response and rejects expired or contradictory login data', () => {
  assert.equal(typeof sessionApi.validateAuthenticatedSession, 'function');
  if (!sessionApi.validateAuthenticatedSession) return;
  assert.doesNotThrow(() => sessionApi.validateAuthenticatedSession!({
    token: token({ user_id: 'root-1', role: 'superadmin', tenant_id: '', exp: 2_000_000_060 }),
    user: { id: 'root-1', role: 'superadmin', tenant_id: '' },
  }));
  assert.throws(() => sessionApi.validateAuthenticatedSession!({
    token: token({ user_id: 'staff-1', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 1 }),
    user: { id: 'staff-1', role: 'tenant_admin', tenant_id: 'tenant-acme' },
    tenant: { tenant_id: 'tenant-acme' },
  }));
  assert.throws(() => sessionApi.validateAuthenticatedSession!({
    token: token({ user_id: 'staff-1', role: 'tenant_admin', tenant_id: 'tenant-bravo', exp: 2_000_000_060 }),
    user: { id: 'staff-1', role: 'tenant_admin', tenant_id: 'tenant-acme' },
    tenant: { tenant_id: 'tenant-acme' },
  }));
});

test('removes a stale portal refresh token when a new session omits it', () => {
  const access = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: 4_000_000_000 });
  const storage = memoryStorage({ [sessionApi.PORTAL_REFRESH_TOKEN_KEY!]: 'stale-refresh' });

  authSession.writePortalSession({ token: access }, 'acme', storage);

  assert.equal(storage.value(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), access);
  assert.equal(storage.value(sessionApi.PORTAL_REFRESH_TOKEN_KEY!), undefined);
});

test('removes a stale admin refresh token when a new staff session omits it', () => {
  const access = token({ user_id: 'staff-1', role: 'tenant_admin', tenant_id: 'tenant-acme', exp: 4_000_000_000 });
  const storage = memoryStorage({ [authSession.ADMIN_REFRESH_TOKEN_KEY]: 'stale-admin-refresh' });

  authSession.writeAdminSession({ token: access }, storage);

  assert.equal(storage.value(authSession.ADMIN_ACCESS_TOKEN_KEY), access);
  assert.equal(storage.value(authSession.ADMIN_REFRESH_TOKEN_KEY), undefined);
});

test('captures both portal tokens for logout before session cleanup', () => {
  const storage = memoryStorage({
    [sessionApi.PORTAL_ACCESS_TOKEN_KEY!]: 'access-before-logout',
    [sessionApi.PORTAL_REFRESH_TOKEN_KEY!]: 'refresh-before-logout',
  });

  assert.deepEqual(sessionApi.createPortalLogoutRequest?.(storage), {
    refreshToken: 'refresh-before-logout',
    authorization: 'Bearer access-before-logout',
  });
});

test('coalesces concurrent portal refreshes and stores the validated learner token once', async () => {
  let resolveRefresh!: (session: { token: string; refresh_token?: string }) => void;
  let requests = 0;
  const refreshed = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: 4_000_000_000 });
  const storage = memoryStorage({
    [sessionApi.PORTAL_ACCESS_TOKEN_KEY!]: token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: 1 }),
    [sessionApi.PORTAL_REFRESH_TOKEN_KEY!]: 'refresh-old',
    [sessionApi.PORTAL_TENANT_CODE_KEY!]: 'acme',
  });
  const request = () => {
    requests += 1;
    return new Promise<{ token: string; refresh_token?: string }>((resolve) => {
      resolveRefresh = resolve;
    });
  };

  assert.equal(typeof sessionApi.createPortalSessionRefresher, 'function');
  if (!sessionApi.createPortalSessionRefresher) return;
  const refresh = sessionApi.createPortalSessionRefresher(
    request,
    () => ({ code: 'acme', tenant_id: 'tenant-acme' }),
    storage,
  );
  const first = refresh();
  const second = refresh();

  assert.equal(first, second);
  assert.equal(requests, 1);
  resolveRefresh({ token: refreshed, refresh_token: 'refresh-new' });
  assert.equal(await first, refreshed);
  assert.equal(storage.value(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), refreshed);
  assert.equal(storage.value(sessionApi.PORTAL_REFRESH_TOKEN_KEY!), 'refresh-new');
});

test('rejects a late refresh response after another login without overwriting the newer session', async () => {
  let resolveRefresh!: (session: { token: string; refresh_token?: string }) => void;
  const oldResult = token({ user_id: 'user-old', role: 'learner', tenant_id: 'tenant-acme', exp: 4_000_000_000 });
  const current = token({ user_id: 'user-new', role: 'learner', tenant_id: 'tenant-acme', exp: 4_000_000_000 });
  const storage = memoryStorage({
    [sessionApi.PORTAL_ACCESS_TOKEN_KEY!]: token({ user_id: 'user-old', role: 'learner', tenant_id: 'tenant-acme', exp: 1 }),
    [sessionApi.PORTAL_REFRESH_TOKEN_KEY!]: 'refresh-old',
    [sessionApi.PORTAL_TENANT_CODE_KEY!]: 'acme',
  });

  assert.equal(typeof sessionApi.createPortalSessionRefresher, 'function');
  if (!sessionApi.createPortalSessionRefresher) return;
  const refresh = sessionApi.createPortalSessionRefresher(
    () => new Promise((resolve) => { resolveRefresh = resolve; }),
    () => ({ code: 'acme', tenant_id: 'tenant-acme' }),
    storage,
  );
  const pending = refresh();
  authSession.writePortalSession({ token: current, refresh_token: 'refresh-current' }, 'acme', storage);
  resolveRefresh({ token: oldResult, refresh_token: 'refresh-from-old-login' });

  await assert.rejects(pending, /superseded|发生变化/);
  assert.equal(storage.value(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), current);
  assert.equal(storage.value(sessionApi.PORTAL_REFRESH_TOKEN_KEY!), 'refresh-current');
});

test('rejects a refreshed learner token for a different portal tenant', async () => {
  const foreign = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-bravo', exp: 4_000_000_000 });
  const oldAccess = token({ user_id: 'user-1', role: 'learner', tenant_id: 'tenant-acme', exp: 1 });
  const storage = memoryStorage({
    [sessionApi.PORTAL_ACCESS_TOKEN_KEY!]: oldAccess,
    [sessionApi.PORTAL_REFRESH_TOKEN_KEY!]: 'refresh-old',
    [sessionApi.PORTAL_TENANT_CODE_KEY!]: 'acme',
  });

  assert.equal(typeof sessionApi.createPortalSessionRefresher, 'function');
  if (!sessionApi.createPortalSessionRefresher) return;
  const refresh = sessionApi.createPortalSessionRefresher(
    async () => ({ token: foreign, refresh_token: 'refresh-foreign' }),
    () => ({ code: 'acme', tenant_id: 'tenant-acme' }),
    storage,
  );

  await assert.rejects(refresh(), /刷新后的企业会话无效/);
  assert.equal(storage.value(sessionApi.PORTAL_ACCESS_TOKEN_KEY!), oldAccess);
  assert.equal(storage.value(sessionApi.PORTAL_REFRESH_TOKEN_KEY!), 'refresh-old');
});

test('refreshes only the first 401 from a protected endpoint when a refresh token exists', () => {
  assert.equal(typeof sessionApi.shouldRefreshPortalRequest, 'function');
  if (!sessionApi.shouldRefreshPortalRequest) return;

  assert.equal(sessionApi.shouldRefreshPortalRequest({
    status: 401,
    url: '/api/v1/courses',
    retried: false,
    hasRefreshToken: true,
  }), true);
  assert.equal(sessionApi.shouldRefreshPortalRequest({
    status: 401,
    url: '/api/v1/courses',
    retried: true,
    hasRefreshToken: true,
  }), false);
  assert.equal(sessionApi.shouldRefreshPortalRequest({
    status: 401,
    url: '/api/v1/auth/login',
    retried: false,
    hasRefreshToken: true,
  }), false);
  assert.equal(sessionApi.shouldRefreshPortalRequest({
    status: 401,
    url: '/api/v1/courses',
    retried: false,
    hasRefreshToken: false,
  }), false);
});
