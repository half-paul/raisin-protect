/**
 * Authentication utilities for Raisin Protect Dashboard
 *
 * Token Strategy (per API spec):
 * - Access token: In-memory (15 min TTL)
 * - Refresh token: localStorage (7 day TTL, single-use rotation)
 * - Auto-refresh on 401 responses
 */

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';

// GRC roles from spec §1.2
export type GrcRole =
  | 'ciso'
  | 'compliance_manager'
  | 'security_engineer'
  | 'it_admin'
  | 'devops_engineer'
  | 'auditor'
  | 'vendor_manager';

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: GrcRole;
  org_id?: string;
  status: 'active' | 'invited' | 'deactivated' | 'locked';
  mfa_enabled?: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at?: string;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  org_name: string;
}

// In-memory token storage
let accessToken: string | null = null;
let tokenExpiresAt: number | null = null;

// State management
type AuthListener = (state: AuthState) => void;
const listeners: Set<AuthListener> = new Set();
let currentState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
};

function notifyListeners() {
  listeners.forEach((listener) => listener(currentState));
}

export function subscribeToAuth(listener: AuthListener): () => void {
  listeners.add(listener);
  listener(currentState);
  return () => listeners.delete(listener);
}

export function getAuthState(): AuthState {
  return currentState;
}

function setTokens(token: string, expiresIn: number, refreshToken?: string) {
  accessToken = token;
  tokenExpiresAt = Date.now() + (expiresIn - 30) * 1000;
  if (refreshToken) {
    localStorage.setItem('rp_refresh_token', refreshToken);
  }
}

function clearTokens() {
  accessToken = null;
  tokenExpiresAt = null;
  localStorage.removeItem('rp_refresh_token');
}

export function getAccessToken(): string | null {
  if (!accessToken || !tokenExpiresAt) return null;
  if (Date.now() >= tokenExpiresAt) {
    accessToken = null;
    tokenExpiresAt = null;
    return null;
  }
  return accessToken;
}

function isTokenExpiringSoon(): boolean {
  if (!tokenExpiresAt) return true;
  return Date.now() >= tokenExpiresAt - 60000;
}

/**
 * Refresh the access token using stored refresh token
 */
export async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = localStorage.getItem('rp_refresh_token');
  if (!refreshToken) return false;

  try {
    const response = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      clearTokens();
      currentState = { user: null, isAuthenticated: false, isLoading: false };
      notifyListeners();
      return false;
    }

    const data = await response.json();
    setTokens(data.data.access_token, data.data.expires_in, data.data.refresh_token);
    return true;
  } catch {
    clearTokens();
    currentState = { user: null, isAuthenticated: false, isLoading: false };
    notifyListeners();
    return false;
  }
}

/**
 * Login with email and password
 */
export async function login(credentials: LoginCredentials): Promise<{
  success: boolean;
  error?: string;
  user?: User;
}> {
  try {
    const response = await fetch(`${API_BASE}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });

    const data = await response.json();

    if (!response.ok) {
      return {
        success: false,
        error: data.error?.message || 'Login failed',
      };
    }

    setTokens(data.data.access_token, data.data.expires_in, data.data.refresh_token);
    const user = data.data.user;
    currentState = { user, isAuthenticated: true, isLoading: false };
    notifyListeners();
    return { success: true, user };
  } catch {
    return { success: false, error: 'Network error' };
  }
}

/**
 * Register a new account and organization
 */
export async function register(credentials: RegisterCredentials): Promise<{
  success: boolean;
  error?: string;
  user?: User;
}> {
  try {
    const response = await fetch(`${API_BASE}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });

    const data = await response.json();

    if (!response.ok) {
      return {
        success: false,
        error: data.error?.message || 'Registration failed',
      };
    }

    setTokens(data.data.access_token, data.data.expires_in, data.data.refresh_token);
    const user = data.data.user;
    currentState = { user, isAuthenticated: true, isLoading: false };
    notifyListeners();
    return { success: true, user };
  } catch {
    return { success: false, error: 'Network error' };
  }
}

/**
 * Logout the current user
 */
export async function logout(): Promise<void> {
  const refreshToken = localStorage.getItem('rp_refresh_token');
  try {
    const token = getAccessToken();
    if (token && refreshToken) {
      await fetch(`${API_BASE}/api/v1/auth/logout`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    }
  } catch {
    // ignore logout errors
  } finally {
    clearTokens();
    currentState = { user: null, isAuthenticated: false, isLoading: false };
    notifyListeners();
  }
}

/**
 * Initialize auth state on app load
 */
export async function initializeAuth(): Promise<void> {
  currentState = { ...currentState, isLoading: true };
  notifyListeners();

  const refreshToken = localStorage.getItem('rp_refresh_token');
  if (!refreshToken) {
    currentState = { user: null, isAuthenticated: false, isLoading: false };
    notifyListeners();
    return;
  }

  const refreshed = await refreshAccessToken();
  if (!refreshed) {
    currentState = { user: null, isAuthenticated: false, isLoading: false };
    notifyListeners();
  }
  // Note: after refresh we have a token but not user data
  // We'll fetch user data via /users endpoint or decode JWT
  // For Sprint 1, we decode the JWT claims for user info
  if (refreshed && accessToken) {
    try {
      const payload = JSON.parse(atob(accessToken.split('.')[1]));
      currentState = {
        user: {
          id: payload.sub,
          email: payload.email,
          first_name: '',
          last_name: '',
          role: payload.role,
          org_id: payload.org,
          status: 'active',
          created_at: '',
        },
        isAuthenticated: true,
        isLoading: false,
      };
      notifyListeners();
    } catch {
      currentState = { user: null, isAuthenticated: false, isLoading: false };
      notifyListeners();
    }
  }
}

/**
 * Make an authenticated API request with auto-refresh on 401
 */
export async function authFetch(
  url: string,
  options: RequestInit = {}
): Promise<Response> {
  let token = getAccessToken();
  if (!token || isTokenExpiringSoon()) {
    const refreshed = await refreshAccessToken();
    if (!refreshed) {
      throw new Error('Not authenticated');
    }
    token = getAccessToken();
  }

  const headers = new Headers(options.headers);
  headers.set('Authorization', `Bearer ${token}`);
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(url, { ...options, headers });

  if (response.status === 401) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      const newToken = getAccessToken();
      headers.set('Authorization', `Bearer ${newToken}`);
      return fetch(url, { ...options, headers });
    }
    clearTokens();
    currentState = { user: null, isAuthenticated: false, isLoading: false };
    notifyListeners();
    throw new Error('Session expired');
  }

  return response;
}

// Role hierarchy helpers
const ADMIN_ROLES: GrcRole[] = ['ciso', 'compliance_manager'];
const USER_MGMT_ROLES: GrcRole[] = ['ciso', 'compliance_manager', 'it_admin'];

export function isAdminRole(role?: GrcRole): boolean {
  return !!role && ADMIN_ROLES.includes(role);
}

export function canManageUsers(role?: GrcRole): boolean {
  return !!role && USER_MGMT_ROLES.includes(role);
}

export function hasRole(user: User | null, ...roles: GrcRole[]): boolean {
  if (!user) return false;
  return roles.includes(user.role);
}

export function getRoleLabel(role: GrcRole): string {
  const labels: Record<GrcRole, string> = {
    ciso: 'CISO / Security Leader',
    compliance_manager: 'GRC / Compliance Manager',
    security_engineer: 'Security Engineer',
    it_admin: 'IT Administrator',
    devops_engineer: 'DevOps Engineer',
    auditor: 'Auditor',
    vendor_manager: 'Vendor Manager',
  };
  return labels[role] || role;
}
