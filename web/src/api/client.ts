/**
 * Typed API client for the Wardrobe backend.
 * Handles credentials, CSRF, abort signals, and error normalization.
 */

import { normalizeApiError } from './errors.js';
import { getCsrfToken, refreshCsrfToken } from './csrf.js';
import { config } from '../app/config.js';

const API_BASE = config.apiBase;

interface FetchOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
}

/**
 * Build the full URL for an API path.
 */
function buildUrl(
  path: string,
  params?: Record<string, string | number | boolean | null | undefined>,
): string {
  const url = new URL(path, API_BASE);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value === null || value === undefined) continue;
      url.searchParams.append(key, String(value));
    }
  }
  return url.toString();
}

/**
 * Serialize request body. Returns body and sets Content-Type if needed.
 */
function serializeBody(body: unknown): { body: BodyInit | undefined; contentType?: string } {
  if (body === null || body === undefined) {
    return { body: undefined };
  }

  if (body instanceof FormData) {
    return { body };
  }

  return { body: JSON.stringify(body), contentType: 'application/json' };
}

/**
 * Main fetch wrapper with CSRF, error handling, and abort support.
 */
export async function apiFetch<T>(
  method: string,
  path: string,
  options: FetchOptions = {},
): Promise<T> {
  const { body: rawBody, headers: extraHeaders, ...restOptions } = options;

  // Serialize body
  const { body, contentType } = serializeBody(rawBody);

  // Build headers
  const headers: Record<string, string> = {};
  if (contentType) {
    headers['Content-Type'] = contentType;
  }
  if (extraHeaders && typeof extraHeaders === 'object' && !Array.isArray(extraHeaders)) {
    for (const [key, value] of Object.entries(extraHeaders)) {
      if (typeof value === 'string') {
        headers[key] = value;
      }
    }
  }

  // Add CSRF for unsafe methods
  if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
    const csrfToken = await getCsrfToken();
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }
  }

  // Build URL with params if present in path
  const url = buildUrl(path);

  // Create abort signal if not provided
  const controller = new AbortController();
  const signal = options.signal || controller.signal;

  try {
    let response = await fetch(url, {
      method,
      credentials: 'same-origin',
      headers,
      body,
      signal,
      ...restOptions,
    });

    // Handle 401 — clear auth and redirect
    if (response.status === 401) {
      handleUnauthorized();
      const error = normalizeApiError(response, null, false);
      throw error;
    }

    // Handle 419 — CSRF expired, refresh and retry once
    if (response.status === 419) {
      await refreshCsrfToken();
      // Retry with new CSRF token
      const newCsrfToken = await getCsrfToken();
      if (newCsrfToken) {
        headers['X-CSRF-Token'] = newCsrfToken;
      }
      response = await fetch(url, {
        method,
        credentials: 'same-origin',
        headers,
        body,
        signal,
        ...restOptions,
      });
    }

    // Parse response
    const contentType = response.headers.get('Content-Type') ?? '';

    // 204 No Content
    if (response.status === 204) {
      return null as T;
    }

    if (!contentType.includes('application/json')) {
      const error = normalizeApiError(response, null, false);
      throw error;
    }

    const responseBody = await response.json();

    if (!response.ok) {
      const error = normalizeApiError(response, responseBody, false);
      throw error;
    }

    return responseBody as T;
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw error;
    }

    if (error instanceof Error && error.name === 'ApiError') {
      throw error;
    }

    // Network error
    const networkError = normalizeApiError(null, null, true);
    throw networkError;
  }
}

/**
 * Unauthorized handler — triggers auth state clear and redirect.
 * Uses a module-level guard to prevent multiple redirects.
 */
let redirectPending = false;

function handleUnauthorized(): void {
  if (redirectPending) return;
  redirectPending = true;

  // Signal to auth store via custom event
  window.dispatchEvent(new CustomEvent('auth:unauthorized'));

  // Redirect after a tick to allow error handling
  setTimeout(() => {
    redirectPending = false;
    const currentPath = window.location.pathname + window.location.search;
    window.location.href = `/login?return_url=${encodeURIComponent(currentPath)}`;
  }, 0);
}

// HTTP method shortcuts
export const apiGet = <T>(
  path: string,
  params?: Record<string, string | number | boolean | null | undefined>,
  options?: RequestInit,
): Promise<T> => {
  const url = buildUrl(path, params);
  return apiFetch<T>('GET', url, options);
};

export const apiPost = <T>(path: string, body?: unknown, options?: RequestInit): Promise<T> =>
  apiFetch<T>('POST', path, { ...options, body });

export const apiPatch = <T>(path: string, body?: unknown, options?: RequestInit): Promise<T> =>
  apiFetch<T>('PATCH', path, { ...options, body });

export const apiPut = <T>(path: string, body?: unknown, options?: RequestInit): Promise<T> =>
  apiFetch<T>('PUT', path, { ...options, body });

export const apiDelete = <T>(path: string, options?: RequestInit): Promise<T> =>
  apiFetch<T>('DELETE', path, { ...options });
