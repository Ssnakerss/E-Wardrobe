/**
 * Application configuration from environment variables.
 * No secrets here — only public settings.
 */

const VITE_API_BASE = import.meta.env.VITE_API_BASE ?? '/api';

export const config = {
  apiBase: VITE_API_BASE,
} as const;
