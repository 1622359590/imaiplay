import assert from 'node:assert/strict'
import test from 'node:test'
import {
  allowedRolesForPath,
  canAccessPath,
  initialOpenGroups,
  navigationForRole,
  pathsForRole,
  roleLabel,
} from '../src/config/adminNavigation.ts'

test('admin navigation starts every page with all groups collapsed', () => {
  const first = initialOpenGroups()
  const second = initialOpenGroups()
  assert.deepEqual(first, [])
  assert.deepEqual(second, [])
  assert.notEqual(first, second)
})

test('instructor navigation exposes only the teaching workspace', () => {
  assert.deepEqual(pathsForRole('instructor'), ['/', '/courses', '/resources'])
})

test('station navigation keeps independent groups open and derives route permissions', () => {
  assert.deepEqual(allowedRolesForPath('/theme-settings'), ['tenant_admin'])
  const openGroups = navigationForRole('tenant_admin').map((group) => group.key)
  assert.deepEqual(openGroups, [
    'home',
    'resource-center',
    'course-center',
    'learner-center',
    'site-settings',
    'security',
  ])
})

test('uses the station role label and denies unknown roles', () => {
  assert.equal(roleLabel('tenant_admin'), '站长')
  assert.deepEqual(pathsForRole('unknown'), [])
  assert.deepEqual(navigationForRole('unknown'), [])
  assert.equal(canAccessPath('tenant_admin', '/theme-settings'), true)
  assert.equal(canAccessPath('instructor', '/theme-settings'), false)
  assert.equal(canAccessPath(undefined, '/'), false)
})
