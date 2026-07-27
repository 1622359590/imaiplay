import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useMemo,
  useState,
} from 'react';
import {
  isAuthenticated,
  login as loginRequest,
  logout as logoutRequest,
  type LoginValues,
} from '../api/auth';

interface AuthContextValue {
  authenticated: boolean;
  login: (values: LoginValues) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [authenticated, setAuthenticated] = useState(isAuthenticated);

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
