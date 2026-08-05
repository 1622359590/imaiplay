import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import {
  legacyPortalRedirect,
  portalLocation,
  portalPath,
  portalRoutePath,
  portalSnapshotForResolution,
  tenantCodeFromPath,
} from '../src/utils/portalRouting.ts';
import * as portalRouting from '../src/utils/portalRouting.ts';
import * as portalSession from '../src/api/portalSession.ts';

const routingApi = portalRouting as typeof portalRouting & {
  portalLoginDestination?: (role: string, tenantCode: string, mode: string) => string;
  boundPortalLoginPath?: (tenantCode?: string) => string | undefined;
  authenticatedLoginDestination?: (
    role: string, tenantCode: string | undefined, mode: string, explicitTenantCode?: string,
  ) => string;
  performLoginNavigation?: (
    destination: string,
    routerNavigate: (destination: string, options: { replace: boolean }) => void,
    documentNavigate: (destination: string) => void,
  ) => void;
  restoredLegacyPortalTarget?: (tenantCode: string, childPath: string) => string;
  portalErrorContent?: (error: unknown) => { title: string; description: string };
};

const portalSessionApi = portalSession as typeof portalSession & {
  requestSessionPortal?: <T>(request: (path: string) => Promise<T>) => Promise<T>;
};

test('extracts the tenant code from a default portal path', () => {
  assert.equal(tenantCodeFromPath('/t/acme/courses'), 'acme');
});

test('only redirects a platform login when a trusted portal code is bound', () => {
  assert.equal(typeof routingApi.boundPortalLoginPath, 'function');
  if (!routingApi.boundPortalLoginPath) return;
  assert.equal(routingApi.boundPortalLoginPath('acme'), '/t/acme');
  assert.equal(routingApi.boundPortalLoginPath(), undefined);
});

test('sends a tenantless superadmin from an explicit portal to platform admin without dereferencing tenant', () => {
  assert.equal(typeof routingApi.authenticatedLoginDestination, 'function');
  if (!routingApi.authenticatedLoginDestination) return;
  assert.equal(
    routingApi.authenticatedLoginDestination('superadmin', undefined, 'default', 'acme'),
    '/admin/',
  );
});

test('rejects a superadmin login on a customer domain before any admin handoff', () => {
  assert.equal(typeof routingApi.authenticatedLoginDestination, 'function');
  if (!routingApi.authenticatedLoginDestination) return;
  assert.throws(
    () => routingApi.authenticatedLoginDestination!('superadmin', undefined, 'custom-domain'),
    /平台管理后台/,
  );
});

test('builds a tenant-safe course path', () => {
  assert.equal(portalPath('acme', '/courses/42'), '/t/acme/courses/42');
});

test('does not infer a tenant from the platform login', () => {
  assert.equal(tenantCodeFromPath('/login'), undefined);
});

test('rejects malformed percent-encoding without throwing', () => {
  assert.equal(tenantCodeFromPath('/t/%E0%A4%A/courses'), undefined);
});

test('preserves the default portal prefix across course navigation', () => {
  assert.equal(portalRoutePath('default', 'acme', '/courses'), '/t/acme/courses');
  assert.equal(
    portalRoutePath('default', 'acme', '/courses/42/lessons/7'),
    '/t/acme/courses/42/lessons/7',
  );
  assert.equal(portalRoutePath('default', 'acme', '/courses/42'), '/t/acme/courses/42');
});

test('keeps course navigation root-relative on a custom domain', () => {
  assert.equal(portalRoutePath('custom-domain', 'acme', '/courses/42'), '/courses/42');
});

test('keeps learner home and recent routes inside each portal while redirecting legacy courses home', () => {
  assert.equal(portalRoutePath('default', 'acme', '/'), '/t/acme');
  assert.equal(portalRoutePath('default', 'acme', '/recent'), '/t/acme/recent');
  assert.equal(portalRoutePath('custom-domain', undefined, '/'), '/');
  assert.equal(portalRoutePath('custom-domain', undefined, '/recent'), '/recent');

  const routerSource = readFileSync(new URL('../src/router.tsx', import.meta.url), 'utf8');
  assert.match(routerSource, /\{ path: 'courses', element: <PortalHomeRedirect \/> \}/);
  assert.match(routerSource, /\{ path: 'recent', element: <RecentPage \/> \}/);
});

test('sends selected learners to their default portal and staff to admin', () => {
  assert.equal(typeof routingApi.portalLoginDestination, 'function');
  if (!routingApi.portalLoginDestination) return;
  assert.equal(routingApi.portalLoginDestination('learner', 'acme', 'default'), '/t/acme');
  assert.equal(routingApi.portalLoginDestination('learner', 'acme', 'custom-domain'), '/');
  assert.equal(routingApi.portalLoginDestination('instructor', 'acme', 'default'), '/admin/');
  assert.equal(routingApi.portalLoginDestination('tenant_admin', 'acme', 'platform'), '/admin/');
  assert.equal(routingApi.portalLoginDestination('superadmin', 'acme', 'platform'), '/admin/');
});

test('uses document navigation for admin handoff and router navigation for learner portals', () => {
  const routerCalls: Array<[string, { replace: boolean }]> = [];
  const documentCalls: string[] = [];
  const routerNavigate = (destination: string, options: { replace: boolean }) => {
    routerCalls.push([destination, options]);
  };
  const documentNavigate = (destination: string) => documentCalls.push(destination);

  assert.equal(typeof routingApi.performLoginNavigation, 'function');
  if (!routingApi.performLoginNavigation) return;
  routingApi.performLoginNavigation('/admin/', routerNavigate, documentNavigate);
  routingApi.performLoginNavigation('/t/acme', routerNavigate, documentNavigate);

  assert.deepEqual(documentCalls, ['/admin/']);
  assert.deepEqual(routerCalls, [['/t/acme', { replace: true }]]);
});

test('resolves the portal for login on a custom host', () => {
  assert.deepEqual(portalLocation('/login', 'academy.example.com'), {
    tenantCode: undefined,
    mode: 'custom-domain',
    shouldResolvePortal: true,
    resolutionKey: 'custom-domain:academy.example.com',
  });
});

test('custom host wins over a conflicting tenant path', () => {
  assert.deepEqual(portalLocation('/t/conflicting/courses', 'academy.example.com'), {
    tenantCode: undefined,
    mode: 'custom-domain',
    shouldResolvePortal: true,
    resolutionKey: 'custom-domain:academy.example.com',
  });
});

test('keeps platform login outside portal resolution', () => {
  assert.deepEqual(portalLocation('/login', 'play.imai.work'), {
    tenantCode: undefined,
    mode: 'platform',
    shouldResolvePortal: false,
    resolutionKey: 'platform',
  });
});

test('hides a previous portal snapshot when the resolution key changes', () => {
  const previous = {
    resolutionKey: 'default:acme',
    portal: { code: 'acme' },
    loading: false,
  };

  assert.deepEqual(
    portalSnapshotForResolution('default:beta', true, previous),
    {
      resolutionKey: 'default:beta',
      loading: true,
    },
  );
});

test('restores a valid learner session for the legacy platform entry', () => {
  assert.deepEqual(legacyPortalRedirect('/pc/courses/42', 'platform', 'learner'), {
    action: 'restore',
    childPath: '/courses/42',
  });
});

test('preserves the legacy child path and session on a custom domain', () => {
  assert.deepEqual(legacyPortalRedirect('/pc/courses', 'custom-domain', 'learner'), {
    action: 'route',
    target: '/courses',
  });
});

test('sends a valid staff legacy session to Admin with document navigation', () => {
  assert.deepEqual(legacyPortalRedirect('/pc/courses', 'platform', 'staff'), {
    action: 'document',
    target: '/admin/',
  });
});

test('sends a legacy platform entry without a valid session to login', () => {
  assert.deepEqual(legacyPortalRedirect('/pc/courses', 'platform', 'none'), {
    action: 'route',
    target: '/login',
  });
});

test('rebuilds the tenant route after restoring the learner portal', () => {
  assert.equal(typeof routingApi.restoredLegacyPortalTarget, 'function');
  assert.equal(
    routingApi.restoredLegacyPortalTarget?.('Acme', '/courses/42'),
    '/t/acme/courses/42',
  );
});

test('loads the authenticated portal from the stable session endpoint', async () => {
  const calls: string[] = [];
  const expected = { code: 'acme', tenant_id: 'tenant-acme' };

  assert.equal(typeof portalSessionApi.requestSessionPortal, 'function');
  if (!portalSessionApi.requestSessionPortal) return;
  const actual = await portalSessionApi.requestSessionPortal(async (path) => {
    calls.push(path);
    return expected;
  });

  assert.deepEqual(calls, ['/api/v1/portal/session']);
  assert.equal(actual, expected);
});

test('maps stable portal error codes to distinct user-facing states', () => {
  assert.equal(typeof routingApi.portalErrorContent, 'function');
  if (!routingApi.portalErrorContent) return;

  assert.deepEqual(
    routingApi.portalErrorContent({ response: { status: 404, data: { error: 'portal_not_found' } } }),
    { title: '门户不存在', description: '请确认门户地址是否正确，或联系企业管理员' },
  );
  assert.deepEqual(
    routingApi.portalErrorContent({ response: { status: 403, data: { error: 'portal_suspended' } } }),
    { title: '租户已暂停', description: '该企业门户已暂停，请联系企业管理员' },
  );
  assert.deepEqual(
    routingApi.portalErrorContent({ response: { status: 403, data: { error: 'portal_trial_expired' } } }),
    { title: '试用已到期', description: '该企业的试用期已结束，请联系企业管理员' },
  );
});
