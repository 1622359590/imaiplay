import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {
  isAuthenticated,
  login as loginRequest,
  logout as logoutRequest,
  persistLogin,
  selectTenant as selectTenantRequest,
  type LoginValues,
  type OrganizationOption,
} from '../api/auth';
import { SESSION_EXPIRED_EVENT } from '../api/authSession';
import type { PortalMode } from '../utils/portalRouting';

export interface PendingOrganizationSelection {
  selectionToken: string;
  organizations: OrganizationOption[];
}

export interface LoginComplete {
  requiresSelection: boolean;
  redirect?: string;
}

interface AuthContextValue {
  authenticated: boolean;
  pendingSelection?: PendingOrganizationSelection;
  login: (values: LoginValues, mode: PortalMode, explicitTenantCode?: string) => Promise<LoginComplete>;
  selectOrganization: (organization: OrganizationOption, mode: PortalMode, explicitTenantCode?: string) => Promise<string>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [authenticated, setAuthenticated] = useState(isAuthenticated);
  const [pendingSelection, setPendingSelection] = useState<PendingOrganizationSelection>();

  useEffect(() => {
    const handleSessionExpired = () => setAuthenticated(false);
    window.addEventListener(SESSION_EXPIRED_EVENT, handleSessionExpired);
    return () => window.removeEventListener(SESSION_EXPIRED_EVENT, handleSessionExpired);
  }, []);

  const login = useCallback(async (
    values: LoginValues,
    mode: PortalMode,
    explicitTenantCode?: string,
  ): Promise<LoginComplete> => {
    const result = await loginRequest(values);
    if (result.requires_tenant_selection) {
      setPendingSelection({ selectionToken: result.selection_token, organizations: result.organizations });
      return { requiresSelection: true };
    }

    const redirect = persistLogin(result, mode, explicitTenantCode);
    setPendingSelection(undefined);
    setAuthenticated(isAuthenticated());
    return { requiresSelection: false, redirect };
  }, []);

  const selectOrganization = useCallback(async (
    organization: OrganizationOption,
    mode: PortalMode,
    explicitTenantCode?: string,
  ): Promise<string> => {
    if (!pendingSelection) throw new Error('请先登录后再选择企业');
    const result = await selectTenantRequest({
      selection_token: pendingSelection.selectionToken,
      tenant_code: organization.code,
    });
    const redirect = persistLogin(result, mode, explicitTenantCode);
    setPendingSelection(undefined);
    setAuthenticated(isAuthenticated());
    return redirect;
  }, [pendingSelection]);

  const logout = useCallback(() => {
    logoutRequest();
    setPendingSelection(undefined);
    setAuthenticated(false);
  }, []);

  const value = useMemo(
    () => ({ authenticated, pendingSelection, login, selectOrganization, logout }),
    [authenticated, pendingSelection, login, selectOrganization, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
}
