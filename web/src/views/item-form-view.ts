/**
 * Item form view (create/edit) - placeholder for Stage 3.
 */

import { element } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';

export class ItemFormView implements View {
  mount(container: HTMLElement): void {
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['New Item'],
      }),
    );
    container.appendChild(
      element({
        tag: 'p',
        children: ['Item form - Stage 3'],
      }),
    );
  }

  destroy(): void {}
}
