/**
 * Confirmation dialog — asks user to confirm an action.
 */

import { element, on } from '../dom/element.js';
import { trapFocus, saveFocus, restoreFocus } from '../dom/focus.js';

interface ConfirmDialogOptions {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  confirmClass?: string;
}

export function createConfirmDialog(opts: ConfirmDialogOptions): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    saveFocus();

    const overlay = element({
      tag: 'div',
      attrs: {
        className: 'dialog-overlay',
        'aria-modal': 'true',
        role: 'dialog',
        'aria-label': opts.title,
      },
    });

    const dialog = element({
      tag: 'dialog',
      className: 'dialog',
      children: [
        element({
          tag: 'h2',
          className: 'dialog__title',
          children: [opts.title],
        }),
        element({
          tag: 'p',
          className: 'dialog__message',
          children: [opts.message],
        }),
        element({
          tag: 'div',
          className: 'dialog__actions',
          children: [
            element({
              tag: 'button',
              attrs: { type: 'button', className: 'btn', 'data-action': 'cancel' },
              children: [opts.cancelLabel ?? 'Cancel'],
            }),
            element({
              tag: 'button',
              attrs: {
                type: 'button',
                className: `btn ${opts.confirmClass ?? 'btn--primary'}`,
                'data-action': 'confirm',
              },
              children: [opts.confirmLabel ?? 'Confirm'],
            }),
          ],
        }),
      ],
    }) as HTMLDialogElement;

    overlay.appendChild(dialog);
    document.body.appendChild(overlay);

    dialog.showModal();

    // Trap focus
    const untrap = trapFocus(dialog);

    const cleanup = () => {
      untrap();
      document.body.removeChild(overlay);
      restoreFocus();
    };

    const handler = (e: Event) => {
      const target = e.target as HTMLElement;
      const action = target.getAttribute('data-action');

      if (action === 'confirm') {
        cleanup();
        resolve(true);
      } else if (action === 'cancel') {
        cleanup();
        resolve(false);
      }
    };

    on(overlay, 'click', (e: Event) => {
      if ((e.target as HTMLElement) === overlay) {
        cleanup();
        resolve(false);
      }
    });

    on(dialog, 'click', handler);
  });
}
