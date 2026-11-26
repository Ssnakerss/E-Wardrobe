/**
 * Focus management utilities.
 */

/**
 * Move focus to an element with visible focus ring.
 */
export function moveFocus(el: HTMLElement | null): void {
  if (!el) return;
  el.focus({ preventScroll: false });
}

/**
 * Trap focus within a container (for dialogs).
 * Returns a cleanup function.
 */
export function trapFocus(container: HTMLElement): () => void {
  const focusableSelectors = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(', ');

  const focusableElements = Array.from(container.querySelectorAll<HTMLElement>(focusableSelectors));

  if (focusableElements.length === 0) {
    return () => {};
  }

  const firstEl = focusableElements.at(0);
  const lastEl = focusableElements.at(-1);

  if (!firstEl || !lastEl) {
    return () => {};
  }

  const handler = (e: KeyboardEvent) => {
    if (e.key !== 'Tab') return;

    if (e.shiftKey && document.activeElement === firstEl) {
      e.preventDefault();
      lastEl.focus();
    } else if (!e.shiftKey && document.activeElement === lastEl) {
      e.preventDefault();
      firstEl.focus();
    }
  };

  document.addEventListener('keydown', handler);
  return () => document.removeEventListener('keydown', handler);
}

/**
 * Get the previously focused element before a dialog/modal opened.
 */
let previousFocus: HTMLElement | null = null;

export function saveFocus(): void {
  previousFocus = document.activeElement as HTMLElement | null;
}

export function restoreFocus(): void {
  if (previousFocus && 'focus' in previousFocus) {
    previousFocus.focus();
    previousFocus = null;
  }
}
