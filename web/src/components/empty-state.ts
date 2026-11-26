/**
 * Empty state component — shown when there is no data to display.
 */

import { element } from '../dom/element.js';

export interface EmptyStateOptions {
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
}

export function createEmptyState(opts: EmptyStateOptions): HTMLElement {
  const children: (Node | string)[] = [
    element({
      tag: 'h2',
      className: 'empty-state__title',
      children: [opts.title],
    }),
  ];

  if (opts.description) {
    children.push(
      element({
        tag: 'p',
        className: 'empty-state__description',
        children: [opts.description],
      }),
    );
  }

  if (opts.actionLabel && opts.onAction) {
    children.push(
      element({
        tag: 'button',
        attrs: {
          type: 'button',
          className: 'btn btn--primary',
        },
        children: [opts.actionLabel],
      }),
    );
  }

  return element({
    tag: 'div',
    attrs: {
      className: 'empty-state',
      role: 'status',
    },
    children,
  });
}
