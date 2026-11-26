/**
 * Not found view - 404 page.
 */

import { element } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';

export class NotFoundView implements View {
  mount(container: HTMLElement): void {
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Page Not Found'],
      }),
    );
    container.appendChild(
      element({
        tag: 'p',
        children: ['The page you are looking for does not exist.'],
      }),
    );
    container.appendChild(
      element({
        tag: 'a',
        attrs: { href: '/', class: 'btn btn--primary' },
        children: ['Go Home'],
      }),
    );
  }

  destroy(): void {}
}
