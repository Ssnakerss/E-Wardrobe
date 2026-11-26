/**
 * Toast notification system.
 */

import { element, on, removeElement } from '../dom/element.js';
import { showToast, dismissToast, subscribeUi } from '../api/ui-store.js';
import type { ToastMessage } from '../api/ui-store.js';

const TOAST_DURATION = 5000;

export function createToastRegion(): HTMLElement {
  const region = element({
    tag: 'div',
    attrs: {
      'aria-live': 'polite',
      'aria-atomic': 'true',
      className: 'toast-region',
    },
  });

  // Listen for toast store changes
  subscribeUi((uiState) => {
    if (uiState.toast) {
      const toast = renderToast(uiState.toast);
      region.appendChild(toast);
      const toastId = uiState.toast.id;
      setTimeout(() => {
        dismissToast(toastId);
        removeToast(toast);
      }, TOAST_DURATION);
    } else if (region.children.length > 0) {
      removeElement(region.firstChild as HTMLElement);
    }
  });

  return region;
}

/**
 * Create and display a toast notification.
 */
export function createToast(message: Omit<ToastMessage, 'id'>): string {
  const id = showToast(message);
  const toast = renderToast({ ...message, id });

  const region = document.querySelector('.toast-region');
  if (region) {
    region.appendChild(toast);
  }

  // Auto-dismiss
  setTimeout(() => {
    dismissToast(id);
    removeToast(toast);
  }, TOAST_DURATION);

  return id;
}

function renderToast(msg: ToastMessage): HTMLElement {
  const toast = element({
    tag: 'div',
    attrs: {
      className: `toast toast--${msg.type}`,
      role: 'alert',
    },
    children: [
      element({
        tag: 'span',
        className: 'toast__message',
        children: [msg.message],
      }),
      element({
        tag: 'button',
        attrs: {
          type: 'button',
          className: 'toast__dismiss',
          'aria-label': 'Dismiss',
          'data-toast-id': msg.id,
        },
        children: ['\u00d7'],
      }),
    ],
  });

  const dismissBtn = toast.querySelector('[data-toast-id]');
  if (dismissBtn) {
    on(dismissBtn, 'click', () => {
      dismissToast(msg.id);
      removeToast(toast);
    });
  }

  return toast;
}

function removeToast(toast: HTMLElement): void {
  const classList = toast.classList;
  classList.add('toast--exiting');
  setTimeout(() => removeElement(toast), 200);
}
