/**
 * Loading state component — shows a spinner while data is being fetched.
 */

import { element } from '../dom/element.js';

export function createLoadingState(message: string = 'Loading...'): HTMLElement {
  return element({
    tag: 'div',
    attrs: {
      className: 'loading-state',
      role: 'status',
      'aria-live': 'polite',
    },
    children: [
      element({
        tag: 'div',
        className: 'loading-state__spinner',
        attrs: { 'aria-hidden': 'true' },
      }),
      element({
        tag: 'p',
        className: 'loading-state__message',
        children: [message],
      }),
    ],
  });
}
