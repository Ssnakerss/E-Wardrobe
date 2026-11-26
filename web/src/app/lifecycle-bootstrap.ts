/**
 * Application lifecycle - bootstraps auth, sets up router, mounts app shell.
 */

import { bootstrapAuth, handleUnauthorized } from '../api/auth-store.js';
import { startRouter } from './router.js';
import { registerRoutes } from './routes.js';
import { createAppHeader } from '../components/app-header.js';
import { createToastRegion } from '../components/toast.js';

export async function bootstrap(): Promise<void> {
  const app = document.getElementById('app');
  if (!app) {
    throw new Error('#app element not found');
  }

  const header = createAppHeader();
  const toastRegion = createToastRegion();

  app.innerHTML = '';
  app.appendChild(header);
  app.appendChild(document.createElement('main'));
  app.appendChild(toastRegion);

  registerRoutes();

  window.addEventListener('auth:unauthorized', () => {
    handleUnauthorized();
  });

  const authState = await bootstrapAuth();

  if (authState.status === 'authenticated') {
    const returnUrl = getReturnUrlFromQuery();
    if (returnUrl && isSafeReturnUrl(returnUrl)) {
      window.location.href = returnUrl;
      return;
    }
    window.location.href = '/items';
    return;
  }

  startRouter();
}

function getReturnUrlFromQuery(): string | null {
  const params = new URLSearchParams(window.location.search);
  return params.get('return_url');
}

function isSafeReturnUrl(url: string): boolean {
  try {
    const parsed = new URL(url, window.location.origin);
    return (
      parsed.origin === window.location.origin &&
      !url.startsWith('http://') &&
      !url.startsWith('https://')
    );
  } catch {
    return false;
  }
}
