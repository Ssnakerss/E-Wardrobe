/**
 * Item details view - placeholder for Stage 3.
 */

import { element } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';

export class ItemDetailsView implements View {
  mount(container: HTMLElement): void {
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Item Details'],
      }),
    );
    container.appendChild(
      element({
        tag: 'p',
        children: ['Item details - Stage 3'],
      }),
    );
  }

  destroy(): void {}
}
