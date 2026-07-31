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
  type LoginValues,
} from '../api/auth';
import { SESSION_EXPIRED_EVENT } from '../api/authSession';

interface AuthContextValue {
  authenticated: boolean;
  login: (values: LoginValues) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [authenticated, setAuthenticated] = useState(isAuthenticated);

  useEffect(() => {
    const handleSessionExpired = () => setAuthenticated(false);
    window.addEventListener(SESSION_EXPIRED_EVENT, handleSessionExpired);
    return () => window.removeEventListener(SESSION_EXPIRED_EVENT, handleSessionExpired);
  }, []);

  const login = useCallback(async (values: LoginValues) => {
    await loginRequest(values);
    setAuthenticated(true);
  }, []);

  const logout = useCallback(() => {
    logoutRequest();
    setAuthenticated(false);
  }, []);

  const value = useMemo(
    () => ({ authenticated, login, logout }),
    [authenticated, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
