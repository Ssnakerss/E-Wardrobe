/**
 * CSRF token management.
 */

interface CsrfState {
  token: string | null;
  fetching: boolean;
}

let state: CsrfState = {
  token: null,
  fetching: false,
};

/**
 * Get the current CSRF token. If none exists, fetch one.
 * Returns the token or null if fetch fails.
 */
export async function getCsrfToken(): Promise<string | null> {
  if (state.token) {
    return state.token;
  }

  if (state.fetching) {
    // Wait for existing fetch — simplified, in production use a promise queue
    return state.token;
  }

  state.fetching = true;
  try {
    const resp = await fetch('/api/v1/auth/csrf', {
      method: 'GET',
      credentials: 'same-origin',
    });

    if (!resp.ok) {
      state.token = null;
      return null;
    }

    const body = await resp.json();
    if (body && typeof body === 'object' && 'csrf_token' in body) {
      state.token = (body as { csrf_token: string }).csrf_token;
      return state.token;
    }

    return null;
  } catch {
    state.token = null;
    return null;
  } finally {
    state.fetching = false;
  }
}

/**
 * Refresh the CSRF token explicitly (called after 401 on CSRF endpoint).
 */
export async function refreshCsrfToken(): Promise<void> {
  state.token = null;
  await getCsrfToken();
}

/**
 * Clear CSRF state on logout.
 */
export function clearCsrfState(): void {
  state.token = null;
  state.fetching = false;
}

/**
 * Check if we have a token.
 */
export function hasCsrfToken(): boolean {
  return state.token !== null;
}
