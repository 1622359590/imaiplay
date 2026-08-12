import { afterEach, describe, expect, it, vi } from 'vitest'
import { deriveClayColors } from '@imaiplay/shared/theme/tenantTheme'

const propertyValues = new Map<string, string>()
const eventListeners = new Map<string, () => void>()
let providerState: unknown
let hasProviderState = false

vi.mock('react', () => ({
  createContext: () => ({ Provider: () => null }),
  useCallback: <T,>(callback: T) => callback,
  useContext: () => null,
  useEffect: (effect: () => void | (() => void)) => { effect() },
  useMemo: <T,>(factory: () => T) => factory(),
  useState: <T,>(initialState: T) => {
    if (!hasProviderState) {
      providerState = initialState
      hasProviderState = true
    }
    return [providerState as T, (nextState: T) => { providerState = nextState }] as const
  },
}))

vi.mock('react-router-dom', () => ({
  useLocation: () => ({ pathname: '/t/acme/login' }),
}))

vi.mock('../src/api/theme', () => ({
  getSessionTenantPortal: vi.fn(),
  getTenantPortal: vi.fn(),
}))

vi.mock('../src/api/authSession', () => ({
  PORTAL_SESSION_EXPIRED_EVENT: 'imaiplay:portal-session-expired',
  bindPortalSession: vi.fn(() => true),
  getActivePortalCode: vi.fn(),
  portalRoutePath: vi.fn((_code: string | undefined, _host: string, childPath: string) => childPath),
  readPortalAccessToken: vi.fn(),
  setActivePortalCode: vi.fn(),
  setActivePortalIdentity: vi.fn(),
}))

vi.mock('../src/api/portalResolution', () => ({
  resolvePortalLocation: () => ({
    key: 'tenant:acme',
    mode: 'tenant',
    shouldResolve: true,
    tenantCode: 'acme',
  }),
  shouldRestoreSessionPortal: () => false,
}))

import { getTenantPortal } from '../src/api/theme'
import { TenantThemeProvider } from '../src/context/TenantThemeContext'

function installDocument() {
  propertyValues.clear()
  eventListeners.clear()
  providerState = undefined
  hasProviderState = false

  vi.stubGlobal('window', {
    location: { hostname: 'learn.example.test' },
    addEventListener: (type: string, listener: () => void) => eventListeners.set(type, listener),
    removeEventListener: (type: string) => eventListeners.delete(type),
    dispatchEvent: vi.fn(),
  })
  vi.stubGlobal('document', {
    documentElement: {
      style: {
        setProperty: (name: string, value: string) => propertyValues.set(name, value),
      },
    },
    title: '',
    querySelector: () => null,
    createElement: () => ({ dataset: {} }),
    head: { appendChild: vi.fn() },
  })
}

async function flushPortalRequest() {
  await Promise.resolve()
  await Promise.resolve()
}

describe('TenantThemeProvider portal palette propagation', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
  })

  it('injects the portal primary and Clay variables into the document root', async () => {
    installDocument()
    const portal = {
      code: 'acme',
      tenant_id: 'tenant-acme',
      name: 'Acme 学习中心',
      primary_color: '#22C55E',
      selected_background_color: '#F0FDF4',
      selected_text_color: '#14532D',
      selected_icon_color: '#166534',
      logo_url: '',
      welcome_text: '',
    }
    const initialClay = deriveClayColors(portal.primary_color)
    const refreshedPrimary = '#8B5CF6'
    const refreshedClay = deriveClayColors(refreshedPrimary)
    expect(refreshedClay.shadow).not.toBe(initialClay.shadow)
    vi.mocked(getTenantPortal)
      .mockResolvedValueOnce(portal)
      .mockResolvedValueOnce(portal)
      .mockResolvedValueOnce({ ...portal, primary_color: refreshedPrimary })
      .mockResolvedValue({ ...portal, primary_color: refreshedPrimary })

    TenantThemeProvider({ children: null })
    await flushPortalRequest()
    TenantThemeProvider({ children: null })

    expect(getTenantPortal).toHaveBeenCalledWith('acme')
    expect(propertyValues.get('--learner-accent')).toBe('#22c55e')
    expect(propertyValues.get('--learner-clay-shadow')).toBe(initialClay.shadow)
    expect(propertyValues.get('--adm-color-primary')).toBe('#22c55e')
    expect(eventListeners.has('tenant-theme-changed')).toBe(true)

    eventListeners.get('tenant-theme-changed')?.()
    await flushPortalRequest()
    TenantThemeProvider({ children: null })

    expect(propertyValues.get('--learner-accent')).toBe('#8b5cf6')
    expect(propertyValues.get('--learner-clay-shadow')).toBe(refreshedClay.shadow)
    expect(propertyValues.get('--adm-color-primary')).toBe('#8b5cf6')
  })
})
