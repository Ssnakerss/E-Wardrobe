/**
 * Version conflict dialog — shows when a 409 conflict occurs.
 */

import { element, on } from '../dom/element.js';
import { saveFocus, restoreFocus } from '../dom/focus.js';
import type { Item, ItemFields } from '../api/items.js';

interface ConflictDialogOptions {
  original: Item;
  draft: Partial<ItemFields>;
  onUseServer: () => void;
  onKeepDraft: () => void;
  onManual: () => void;
}

export function createConflictDialog(opts: ConflictDialogOptions): void {
  saveFocus();

  const overlay = element({
    tag: 'div',
    attrs: {
      className: 'dialog-overlay',
      'aria-modal': 'true',
      role: 'dialog',
      'aria-label': 'Version conflict',
    },
  });

  const changedFields = Object.entries(opts.draft).filter(([key, value]) => {
    const originalValue = (opts.original as unknown as Record<string, unknown>)[key];
    return originalValue !== value;
  });

  const changesList = element({
    tag: 'div',
    className: 'conflict-dialog__changes',
  });

  for (const [key, value] of changedFields) {
    const label = key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
    const row = element({
      tag: 'div',
      className: 'conflict-dialog__change',
      children: [
        element({ tag: 'strong', children: [`${label}:`] }),
        element({
          tag: 'span',
          className: 'conflict-dialog__value',
          children: [String(value ?? '—')],
        }),
      ],
    });
    changesList.appendChild(row);
  }

  const dialog = element({
    tag: 'dialog',
    className: 'dialog dialog--conflict',
    children: [
      element({
        tag: 'h2',
        className: 'dialog__title',
        children: ['Version conflict detected'],
      }),
      element({
        tag: 'p',
        className: 'dialog__message',
        children: ['Someone else modified this item. Choose how to resolve:'],
      }),
      changesList,
      element({
        tag: 'div',
        className: 'dialog__actions dialog__actions--column',
        children: [
          element({
            tag: 'button',
            attrs: { type: 'button', className: 'btn', 'data-action': 'server' },
            children: ['Load server version (discard changes)'],
          }),
          element({
            tag: 'button',
            attrs: { type: 'button', className: 'btn btn--primary', 'data-action': 'draft' },
            children: ['Keep my changes'],
          }),
          element({
            tag: 'button',
            attrs: { type: 'button', className: 'btn', 'data-action': 'manual' },
            children: ['Review changes manually'],
          }),
        ],
      }),
    ],
  }) as HTMLDialogElement;

  overlay.appendChild(dialog);
  document.body.appendChild(overlay);

  dialog.showModal();

  const handler = (e: Event) => {
    const target = e.target as HTMLElement;
    const action = target.getAttribute('data-action');

    if (action === 'server') {
      document.body.removeChild(overlay);
      restoreFocus();
      opts.onUseServer();
    } else if (action === 'draft') {
      document.body.removeChild(overlay);
      restoreFocus();
      opts.onKeepDraft();
    } else if (action === 'manual') {
      document.body.removeChild(overlay);
      restoreFocus();
      opts.onManual();
    }
  };

  on(overlay, 'click', (e: Event) => {
    if ((e.target as HTMLElement) === overlay) {
      document.body.removeChild(overlay);
      restoreFocus();
    }
  });

  on(dialog, 'click', handler);
}
