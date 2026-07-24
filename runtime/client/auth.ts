/**
 * Zyra Client Authentication & Authorization Hook (`useZyraAuth`)
 * Provides state management and helper methods for authentication & RBAC.
 */

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

export interface ZyraUser {
  id: string;
  email: string;
  emailVerified: boolean;
  twoFactorEnabled: boolean;
  roles: string[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

export interface RegisterInput {
  email: string;
  password: string;
}

export interface LoginInput {
  email: string;
  password: string;
  totpCode?: string;
}

export interface ResetPasswordInput {
  token: string;
  newPassword: string;
}

export interface TOTPSetupResponse {
  secret: string;
  qrCodeUrl: string;
}

export interface ZyraAuthContextType {
  user: ZyraUser | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  login: (input: LoginInput) => Promise<ZyraUser>;
  logout: () => Promise<void>;
  register: (input: RegisterInput) => Promise<{ user: ZyraUser; verificationToken: string }>;
  verifyEmail: (token: string) => Promise<void>;
  resetPassword: (input: ResetPasswordInput) => Promise<void>;
  setup2FA: () => Promise<TOTPSetupResponse>;
  verify2FA: (code: string) => Promise<void>;
  refreshUser: () => Promise<ZyraUser | null>;
  hasRole: (role: string) => boolean;
  hasPermission: (permission: string) => boolean;
}

const ZyraAuthContext = createContext<ZyraAuthContextType | undefined>(undefined);

export function ZyraAuthProvider({ children }: { children: ReactNode }) {
  const auth = useProvideZyraAuth();
  return React.createElement(ZyraAuthContext.Provider, { value: auth }, children);
}

export function useZyraAuth(): ZyraAuthContextType {
  const context = useContext(ZyraAuthContext);
  if (!context) {
    // Return standalone fallback instance if provider is omitted
    return useProvideZyraAuth();
  }
  return context;
}

function useProvideZyraAuth(): ZyraAuthContextType {
  const [user, setUser] = useState<ZyraUser | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const getCsrfToken = (): string => {
    if (typeof document === 'undefined') return '';
    const match = document.cookie.match(new RegExp('(^| )_zyra_csrf=([^;]+)'));
    return match ? match[2] : '';
  };

  const refreshUser = async (): Promise<ZyraUser | null> => {
    setLoading(true);
    try {
      const res = await fetch('/_zyra/auth/me', {
        headers: { 'X-CSRF-Token': getCsrfToken() },
      });
      if (res.ok) {
        const json = await res.json();
        if (json.ok && json.data) {
          setUser(json.data);
          setError(null);
          return json.data;
        }
      }
      setUser(null);
      return null;
    } catch (err: any) {
      setUser(null);
      return null;
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refreshUser();
  }, []);

  const login = async (input: LoginInput): Promise<ZyraUser> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/_zyra/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify(input),
      });

      const json = await res.json();
      if (!json.ok) {
        throw new Error(json.error?.message || 'Login failed');
      }

      setUser(json.data.user);
      return json.data.user;
    } catch (err: any) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    setLoading(true);
    try {
      await fetch('/_zyra/auth/logout', {
        method: 'POST',
        headers: { 'X-CSRF-Token': getCsrfToken() },
      });
    } finally {
      setUser(null);
      setLoading(false);
    }
  };

  const register = async (input: RegisterInput): Promise<{ user: ZyraUser; verificationToken: string }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/_zyra/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify(input),
      });

      const json = await res.json();
      if (!json.ok) {
        throw new Error(json.error?.message || 'Registration failed');
      }

      return json.data;
    } catch (err: any) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const verifyEmail = async (token: string): Promise<void> => {
    setLoading(true);
    try {
      const res = await fetch('/_zyra/auth/verify-email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify({ token }),
      });

      const json = await res.json();
      if (!json.ok) {
        throw new Error(json.error?.message || 'Email verification failed');
      }
    } catch (err: any) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const resetPassword = async (input: ResetPasswordInput): Promise<void> => {
    setLoading(true);
    try {
      const res = await fetch('/_zyra/auth/reset-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify(input),
      });

      const json = await res.json();
      if (!json.ok) {
        throw new Error(json.error?.message || 'Password reset failed');
      }
    } catch (err: any) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const setup2FA = async (): Promise<TOTPSetupResponse> => {
    const res = await fetch('/_zyra/auth/2fa/setup', {
      method: 'POST',
      headers: { 'X-CSRF-Token': getCsrfToken() },
    });
    const json = await res.json();
    if (!json.ok) throw new Error(json.error?.message || 'Failed to setup 2FA');
    return json.data;
  };

  const verify2FA = async (code: string): Promise<void> => {
    const res = await fetch('/_zyra/auth/2fa/verify', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': getCsrfToken(),
      },
      body: JSON.stringify({ code }),
    });
    const json = await res.json();
    if (!json.ok) throw new Error(json.error?.message || '2FA code verification failed');
  };

  const hasRole = (role: string): boolean => {
    if (!user) return false;
    return user.roles.includes(role) || user.roles.includes('admin');
  };

  const hasPermission = (permission: string): boolean => {
    if (!user) return false;
    if (user.permissions.includes('*')) return true;
    return user.permissions.includes(permission);
  };

  return {
    user,
    isAuthenticated: !!user,
    loading,
    error,
    login,
    logout,
    register,
    verifyEmail,
    resetPassword,
    setup2FA,
    verify2FA,
    refreshUser,
    hasRole,
    hasPermission,
  };
}

export interface ZyraProtectedRouteProps {
  role?: string;
  permission?: string;
  fallback?: ReactNode;
  children: ReactNode;
}

export function ZyraProtectedRoute({ role, permission, fallback = null, children }: ZyraProtectedRouteProps) {
  const { isAuthenticated, loading, hasRole, hasPermission } = useZyraAuth();

  if (loading) {
    return React.createElement('div', { className: 'zyra-auth-loading' }, 'Loading...');
  }

  if (!isAuthenticated) {
    return fallback ? React.createElement(React.Fragment, null, fallback) : null;
  }

  if (role && !hasRole(role)) {
    return fallback ? React.createElement(React.Fragment, null, fallback) : null;
  }

  if (permission && !hasPermission(permission)) {
    return fallback ? React.createElement(React.Fragment, null, fallback) : null;
  }

  return React.createElement(React.Fragment, null, children);
}
