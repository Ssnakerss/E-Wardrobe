/**
 * Categories view — delegates to category-nav component.
 */

import { createCategoryNav } from '../components/category-nav.js';
import type { View } from '../app/lifecycle.js';

export class CategoriesView implements View {
  private navView: ReturnType<typeof createCategoryNav> | null = null;

  mount(container: HTMLElement): void {
    this.navView = createCategoryNav({
      onCategorySelect: (_categoryId: string) => {
        // Stage 3: navigate to items filtered by category
        console.log('Category selected:', _categoryId);
      },
    });

    this.navView.mount(container);
  }

  destroy(): void {
    this.navView?.destroy();
    this.navView = null;
  }
}
