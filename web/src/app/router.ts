/**
 * Simple History API router with route guards and view lifecycle.
 */

import { on } from '../dom/element.js';
import type { View } from './lifecycle.js';

export interface Route {
  path: string;
  pattern: RegExp;
  paramNames: string[];
  guard: () => boolean; // true = allowed
  load: () => Promise<View>;
}

const routes: Route[] = [];
let currentView: View | null = null;
let currentParams: Record<string, string> = {};

/**
 * Register a route.
 */
export function addRoute(path: string, guard: () => boolean, load: () => Promise<View>): void {
  const paramNames: string[] = [];
  const pattern = path
    .replace(/:(\w+)/g, (_match, name) => {
      paramNames.push(name);
      return '([^/]+)';
    })
    .replace(/\./g, '\\.');

  routes.push({
    path,
    pattern: new RegExp(`^${pattern}$`),
    paramNames,
    guard,
    load,
  });
}

/**
 * Parse a URL path into route match.
 */
function matchRoute(pathname: string): { route: Route; params: Record<string, string> } | null {
  for (const route of routes) {
    const match = pathname.match(route.pattern);
    if (match) {
      const params: Record<string, string> = {};
      route.paramNames.forEach((name, i) => {
        const value = match[i + 1];
        if (value !== undefined) {
          params[name] = value;
        }
      });
      return { route, params };
    }
  }
  return null;
}

/**
 * Navigate to a path.
 */
export function navigate(path: string, replace?: boolean): void {
  if (replace) {
    window.history.replaceState({ path }, '', path);
  } else {
    window.history.pushState({ path }, '', path);
  }
  handlePopState();
}

/**
 * Handle browser back/forward.
 */
async function handlePopState(): Promise<void> {
  const pathname = window.location.pathname + window.location.search;
  const matched = matchRoute(pathname);

  // Destroy previous view
  if (currentView) {
    currentView.destroy();
    currentView = null;
  }

  if (matched) {
    if (!matched.route.guard()) {
      // Redirect to login
      const currentPath = window.location.pathname + window.location.search;
      navigate(`/login?return_url=${encodeURIComponent(currentPath)}`, true);
      return;
    }

    currentParams = matched.params;
    try {
      const view = await matched.route.load();
      currentView = view;

      // Mount view to main element
      const main = document.querySelector('main');
      if (main) {
        main.innerHTML = '';
        await view.mount(main);
      }
    } catch (error) {
      console.error('Failed to load view:', error);
      const main = document.querySelector('main');
      if (main) {
        main.innerHTML = '<p>Failed to load page.</p>';
      }
    }
  } else {
    // 404 — dispatch event for view to handle
    window.dispatchEvent(new CustomEvent('router:404', { detail: pathname }));
  }
}

/**
 * Start the router.
 */
export function startRouter(): void {
  // Intercept link clicks
  on(
    document.body,
    'click',
    (e: Event) => {
      const mouseEvent = e as MouseEvent;
      const target = (mouseEvent.target as HTMLElement).closest('a');
      if (!target) return;

      const href = target.getAttribute('href');
      if (!href) return;

      // Ignore non-same-origin, modifier clicks, downloads
      try {
        if (mouseEvent.ctrlKey || mouseEvent.metaKey || mouseEvent.altKey || mouseEvent.shiftKey)
          return;
        if (mouseEvent.button !== 0) return;
        if (target.hasAttribute('download')) return;

        const parsedUrl = new URL(href, window.location.origin);
        if (parsedUrl.origin !== window.location.origin) return;

        mouseEvent.preventDefault();
        navigate(parsedUrl.pathname + parsedUrl.search);
      } catch {
        return;
      }
    },
    true,
  );

  // Handle popstate
  on(window, 'popstate', () => {
    void handlePopState();
  });

  // Initial navigation
  void handlePopState();
}

/**
 * Get current route params.
 */
export function getCurrentParams(): Record<string, string> {
  return currentParams;
}

/**
 * Get current pathname.
 */
export function getCurrentPath(): string {
  return window.location.pathname + window.location.search;
}
