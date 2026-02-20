'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import {
  AuthState,
  User,
  GrcRole,
  LoginCredentials,
  RegisterCredentials,
  subscribeToAuth,
  initializeAuth,
  login as authLogin,
  logout as authLogout,
  register as authRegister,
  hasRole as checkHasRole,
  isAdminRole,
  canManageUsers as checkCanManageUsers,
} from './auth';

interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  register: (credentials: RegisterCredentials) => Promise<{ success: boolean; error?: string }>;
  hasRole: (...roles: GrcRole[]) => boolean;
  isAdmin: boolean;
  canManageUsers: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
  });

  useEffect(() => {
    const unsubscribe = subscribeToAuth(setState);
    initializeAuth();
    return unsubscribe;
  }, []);

  const login = useCallback(async (credentials: LoginCredentials) => {
    const result = await authLogin(credentials);
    return { success: result.success, error: result.error };
  }, []);

  const logout = useCallback(async () => {
    await authLogout();
  }, []);

  const register = useCallback(async (credentials: RegisterCredentials) => {
    const result = await authRegister(credentials);
    return { success: result.success, error: result.error };
  }, []);

  const hasRole = useCallback(
    (...roles: GrcRole[]) => checkHasRole(state.user, ...roles),
    [state.user]
  );

  const value: AuthContextType = {
    ...state,
    login,
    logout,
    register,
    hasRole,
    isAdmin: isAdminRole(state.user?.role),
    canManageUsers: checkCanManageUsers(state.user?.role),
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
