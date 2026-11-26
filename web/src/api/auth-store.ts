/**
 * Auth store - manages user state and CSRF lifecycle.
 */

import { clearCsrfState } from './csrf.js';

export interface User {
  id: string;
  email: string;
  status: 'active' | 'disabled';
}

export type AuthState =
  | { status: 'loading' }
  | { status: 'authenticated'; user: User }
  | { status: 'unauthenticated' }
  | { status: 'error'; message: string };

type AuthListener = (state: AuthState) => void;

let state: AuthState = { status: 'loading' };
const listeners: Set<AuthListener> = new Set();

export function getAuthState(): AuthState {
  return state;
}

export function subscribeAuth(listener: AuthListener): () => void {
  listeners.add(listener);
  listener(state);
  return () => {
    listeners.delete(listener);
  };
}

function setState(newState: AuthState): void {
  state = newState;
  for (const listener of listeners) {
    listener(state);
  }
}

export async function bootstrapAuth(): Promise<AuthState> {
  try {
    const resp = await fetch('/api/v1/auth/me', {
      credentials: 'same-origin',
    });

    if (!resp.ok) {
      throw new Error('Not authenticated');
    }

    const data = await resp.json();

    if (data && typeof data === 'object' && 'id' in data && 'email' in data && 'status' in data) {
      const user = data as User;
      setState({ status: 'authenticated', user });
      return { status: 'authenticated', user };
    }
  } catch {
    // Not authenticated or network error
  }

  setState({ status: 'unauthenticated' });
  return { status: 'unauthenticated' };
}

export function setAuthenticated(user: User): void {
  setState({ status: 'authenticated', user });
}

export function clearAuth(): void {
  clearCsrfState();
  setState({ status: 'unauthenticated' });
}

export function handleUnauthorized(): void {
  clearAuth();
}

export function isAuthenticated(): boolean {
  return state.status === 'authenticated';
}

export function getCurrentUser(): User | null {
  if (state.status === 'authenticated') {
    return state.user;
  }
  return null;
}
