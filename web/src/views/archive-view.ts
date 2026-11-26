/**
 * Archive view - placeholder for Stage 3.
 */

import { element } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';

export class ArchiveView implements View {
  mount(container: HTMLElement): void {
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Archive'],
      }),
    );
    container.appendChild(
      element({
        tag: 'p',
        children: ['Archive - Stage 3'],
      }),
    );
  }

  destroy(): void {}
}
