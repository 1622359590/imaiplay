import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react';
import { useLocation } from 'react-router-dom';
import { resolvePortal, type Portal } from '../api/portal';
import {
  bindPortalSessionToPortal,
  readPortalAccessToken,
  SESSION_EXPIRED_EVENT,
} from '../api/authSession';
import {
  clearActivePortalCode,
  setActivePortalCode,
  setActivePortalIdentity,
} from '../api/portalSession';
import {
  portalLocation,
  portalSnapshotForResolution,
  type PortalMode,
  type PortalSnapshot,
} from '../utils/portalRouting';

const DEFAULT_FAVICON = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='16' fill='%232563EB'/%3E%3Cpath d='M20 18h24v28H20z' fill='none' stroke='white' stroke-width='4'/%3E%3Cpath d='M26 25h12M26 32h12M26 39h8' stroke='white' stroke-width='3'/%3E%3C/svg%3E";

function applyPortalBranding(portal?: Portal) {
  document.title = portal?.browser_title?.trim() || (portal ? `${portal.name} | 企业学习中心` : 'iMaiPlay 企业学习中心');
  let favicon = document.querySelector<HTMLLinkElement>('link[data-imaiplay-favicon]');
  if (!favicon) {
    favicon = document.createElement('link');
    favicon.rel = 'icon';
    favicon.dataset.imaiplayFavicon = 'true';
    document.head.appendChild(favicon);
  }
  favicon.href = portal?.logo_url || DEFAULT_FAVICON;
}

export type { PortalMode } from '../utils/portalRouting';

export interface PortalContextValue {
  portal?: Portal;
  tenantCode?: string;
  mode: PortalMode;
  loading: boolean;
  error?: unknown;
}

const PortalContext = createContext<PortalContextValue | null>(null);

export function PortalProvider({ children }: PropsWithChildren) {
  const location = useLocation();
  const {
    tenantCode,
    mode,
    shouldResolvePortal,
    resolutionKey,
  } = portalLocation(location.pathname, window.location.hostname);
  const [state, setState] = useState<PortalSnapshot<Portal>>({
    resolutionKey,
    loading: shouldResolvePortal,
  });

  useEffect(() => {
    let cancelled = false;

    if (!shouldResolvePortal) {
      clearActivePortalCode();
      setState({ resolutionKey, loading: false });
      return () => { cancelled = true; };
    }

    if (tenantCode) setActivePortalCode(tenantCode);
    setState({ resolutionKey, loading: true });
    void resolvePortal(tenantCode)
      .then((portal) => {
        if (cancelled) return;
        const hadPortalSession = Boolean(readPortalAccessToken());
        if (hadPortalSession && !bindPortalSessionToPortal(portal)) {
          window.dispatchEvent(new Event(SESSION_EXPIRED_EVENT));
        }
        setActivePortalIdentity(portal);
        applyPortalBranding(portal);
        setState({ resolutionKey, portal, loading: false });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        clearActivePortalCode();
        setState({ resolutionKey, loading: false, error });
      });

    return () => { cancelled = true; };
  }, [resolutionKey, shouldResolvePortal, tenantCode]);

  const currentState = portalSnapshotForResolution(
    resolutionKey,
    shouldResolvePortal,
    state,
  );
  const value = useMemo<PortalContextValue>(
    () => ({
      portal: currentState.portal,
      tenantCode: currentState.portal?.code ?? tenantCode,
      mode,
      loading: currentState.loading,
      error: currentState.error,
    }),
    [currentState, mode, tenantCode],
  );

  return <PortalContext.Provider value={value}>{children}</PortalContext.Provider>;
}

export function usePortal(): PortalContextValue {
  const context = useContext(PortalContext);
  if (!context) {
    throw new Error('usePortal must be used within PortalProvider');
  }
  return context;
}
