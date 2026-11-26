/**
 * Categories view - placeholder for Stage 2.
 */

import { element } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';

export class CategoriesView implements View {
  mount(container: HTMLElement): void {
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Categories'],
      }),
    );
    container.appendChild(
      element({
        tag: 'p',
        children: ['Categories - Stage 2'],
      }),
    );
  }

  destroy(): void {}
}
