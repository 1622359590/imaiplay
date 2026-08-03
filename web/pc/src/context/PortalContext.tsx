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
