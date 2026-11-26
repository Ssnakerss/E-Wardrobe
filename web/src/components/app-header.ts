/**
 * App header with navigation and user info.
 */

import { element, on } from '../dom/element.js';
import { getAuthState } from '../api/auth-store.js';
import { navigate } from '../app/router.js';

export function createAppHeader(): HTMLElement {
  const header = element({
    tag: 'header',
    className: 'app-header',
  });

  render(header);

  // Re-render on auth state change
  // For simplicity, re-render is triggered by navigation events
  // A full implementation would subscribe to auth store

  return header;
}

function render(header: HTMLElement): void {
  const authState = getAuthState();

  header.innerHTML = '';

  const nav = element({
    tag: 'nav',
    attrs: { 'aria-label': 'Main navigation' },
  });

  // Logo / brand
  const brand = element({
    tag: 'a',
    attrs: {
      href: authState.status === 'authenticated' ? '/items' : '/',
      className: 'app-brand',
    },
    children: [
      element({
        tag: 'h1',
        attrs: { class: 'app-brand-text' },
        children: ['E-Wardrobe'],
      }),
    ],
  });

  nav.appendChild(brand);

  // Navigation links
  const links = element({ tag: 'ul', className: 'nav-links' });

  if (authState.status === 'authenticated') {
    const itemsLink = element({
      tag: 'li',
      children: [
        element({
          tag: 'a',
          attrs: { href: '/items', className: 'nav-link' },
          children: ['Items'],
        }),
      ],
    });

    const categoriesLink = element({
      tag: 'li',
      children: [
        element({
          tag: 'a',
          attrs: { href: '/categories', className: 'nav-link' },
          children: ['Categories'],
        }),
      ],
    });

    const archiveLink = element({
      tag: 'li',
      children: [
        element({
          tag: 'a',
          attrs: { href: '/archive', className: 'nav-link' },
          children: ['Archive'],
        }),
      ],
    });

    const logoutBtn = element({
      tag: 'li',
      children: [
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: 'nav-link nav-link--danger',
            'data-action': 'logout',
          },
          children: ['Logout'],
        }),
      ],
    });

    links.append(itemsLink, categoriesLink, archiveLink, logoutBtn);
  } else {
    const loginLink = element({
      tag: 'li',
      children: [
        element({
          tag: 'a',
          attrs: { href: '/login', className: 'nav-link' },
          children: ['Login'],
        }),
      ],
    });

    links.appendChild(loginLink);
  }

  nav.appendChild(links);
  header.appendChild(nav);

  // Logout handler
  if (authState.status === 'authenticated') {
    const logoutBtn = header.querySelector('[data-action="logout"]');
    if (logoutBtn) {
      on(logoutBtn as EventTarget, 'click', async () => {
        try {
          const { logout } = await import('../api/auth.js');
          const { clearAuth } = await import('../api/auth-store.js');
          await logout();
          clearAuth();
          navigate('/login');
        } catch {
          // On error, clear auth anyway
          const { clearAuth } = await import('../api/auth-store.js');
          clearAuth();
          navigate('/login');
        }
      });
    }
  }
}
