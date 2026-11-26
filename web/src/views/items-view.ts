/**
 * Items list view - placeholder for Stage 3.
 */

import { element } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';

export class ItemsView implements View {
  mount(container: HTMLElement): void {
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Items'],
      }),
    );
    container.appendChild(
      element({
        tag: 'p',
        children: ['Items list - Stage 3'],
      }),
    );
  }

  destroy(): void {}
}
