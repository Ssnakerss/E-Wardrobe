/**
 * Error state component — shown when an API call or operation fails.
 */

import { element } from '../dom/element.js';
import type { ApiError } from '../api/errors.js';

export interface ErrorStateOptions {
  title?: string;
  message: string;
  retryLabel?: string;
  onRetry?: () => void;
}

export function createErrorState(opts: ErrorStateOptions): HTMLElement {
  const children: (Node | string)[] = [
    element({
      tag: 'h2',
      className: 'error-state__title',
      children: [opts.title ?? 'Error'],
    }),
    element({
      tag: 'p',
      className: 'error-state__message',
      children: [opts.message],
    }),
  ];

  if (opts.retryLabel && opts.onRetry) {
    children.push(
      element({
        tag: 'button',
        attrs: {
          type: 'button',
          className: 'btn btn--primary',
        },
        children: [opts.retryLabel],
      }),
    );
  }

  return element({
    tag: 'div',
    attrs: {
      className: 'error-state',
      role: 'alert',
    },
    children,
  });
}

/**
 * Create an error state from an ApiError.
 */
export function errorStateFromApiError(
  error: ApiError,
  retryLabel?: string,
  onRetry?: () => void,
): HTMLElement {
  const title = getErrorTitle(error.kind);
  return createErrorState({
    title,
    message: error.message,
    retryLabel,
    onRetry,
  });
}

function getErrorTitle(kind: string): string {
  switch (kind) {
    case 'network':
      return 'Connection error';
    case 'unauthorized':
      return 'Session expired';
    case 'forbidden':
      return 'Access denied';
    case 'not_found':
      return 'Not found';
    case 'version_conflict':
      return 'Conflict detected';
    case 'payload_too_large':
      return 'File too large';
    case 'validation':
      return 'Validation error';
    case 'rate_limited':
      return 'Too many requests';
    case 'server':
      return 'Server error';
    default:
      return 'Error';
  }
}
