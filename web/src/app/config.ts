/**
 * Application configuration from environment variables.
 * No secrets here — only public settings.
 */

const VITE_API_BASE = import.meta.env.VITE_API_BASE;

/**
 * Returns the API base URL.
 * In production (Vite build) VITE_API_BASE is undefined, so we default to '/api'.
 * We always return an absolute URL because new URL(path, base) requires an absolute base.
 */
function getApiBase(): string {
  if (VITE_API_BASE) return VITE_API_BASE;
  // Fallback: use current origin + default path
  return `${window.location.protocol}//${window.location.host}/api`;
}

export const config = {
  apiBase: getApiBase(),
} as const;
